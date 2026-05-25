# Gutter — Roadmap triển khai

Kế hoạch kỹ thuật chi tiết để gỡ trần use-case (cold-start/SEO) và đào sâu moat
"Go full-stack", dựa trên kết quả benchmark trong `bench/ANALYSIS.md`.

**Chẩn đoán từ số liệu:** runtime đã đủ tốt (tương tác hòa, compute hòa/thắng).
Điểm thua duy nhất là **chi phí vào cửa**: bundle WASM + ~15–20ms instantiate +
không SSR/SEO/a11y. Vì vậy mọi phase dưới đây nhắm vào: (1) xóa chi phí vào cửa,
(2) khóa chặt lợi thế một-ngôn-ngữ-cả-stack.

**Nguyên tắc nền (giữ nguyên qua mọi phase):**
- Widget là dữ liệu thuần (`Widget = any`, `Host()` trả struct). Mọi thứ chạm
  `syscall/js` phải nằm trong file `*_wasm.go` + có `*_stub.go` đối ứng.
- Mỗi tính năng người dùng thấy được phải có test ở đúng layer (xem `TESTING.md`):
  layer-1 host unit, layer-2 reconciler-vs-DOM, layer-3 Playwright e2e.
- Không phá API hiện có; thêm là chính, đổi thì giữ tương thích ngược.

Quy ước effort: **S** ≈ vài ngày · **M** ≈ 1–2 tuần · **L** ≈ 3–5 tuần · **XL** ≈ 6+ tuần.

---

## Phase 0 — Quả ngọt hái liền (mở đường, đo baseline)

Mục tiêu: giảm ngay chi phí vào cửa và dựng hạ tầng đo để chứng minh các phase sau.

### 0.1 Production mặc định TinyGo  — **S**
- **Vì sao:** TinyGo cắt bundle 792KB→306KB gz và cold render ~⅓ (số liệu mục 1–2),
  gần như miễn phí về công.
- **Việc làm:** trong `cmd/gutter/build.go`, thêm cờ `build deploy` (production)
  mặc định `--tinygo` khi có tinygo; fallback Go-std + cảnh báo. Giữ `build`/`run`
  dev là Go-std (compile nhanh).
- **Cảnh báo cần ghi tài liệu:** TinyGo phình RAM ở cây cực lớn (mục 6) → chỉ bật
  mặc định cho production, document rõ.
- **Test:** layer-3 — thêm CI job build testapp bằng TinyGo + chạy e2e hiện có.
- **Done khi:** `gutter build deploy` ra bundle TinyGo chạy qua trọn bộ e2e.

### 0.2 Accessibility nền  — **M**
- **Vì sao:** `Heading` render `<span>`, không ARIA → chặn app nghiêm túc & SEO.
- **Việc làm:**
  - `widgets/typography.go`: `Heading{Level}` → tag `<h1>..<h6>` thật (giữ style).
  - Thêm field `Role`/`AriaLabel` vào `Styled` & các widget tương tác; `Button`,
    `IconButton`, `Switch` (đã là `role=switch`) phát đúng ARIA.
  - `Link` → `<a href>` thật; `Image` đã có `<img>`, thêm `Alt`.
- **Test:** layer-1 — assert `hostOf(...)` ra đúng tag/role; layer-3 — `getByRole('heading')`
  (đang trống, xem "Known limitations") phải tìm thấy.
- **Done khi:** e2e role-query xanh; không vỡ snapshot style hiện tại.

### 0.3 Tích hợp benchmark vào CI làm "performance gate"  — **S**
- **Việc làm:** đưa `bench/` thành job đo định kỳ (không chặn PR), lưu
  `RESULTS.md/RESOURCES.md/COMPUTE.md` làm artifact để theo dõi hồi quy.
- **Done khi:** mỗi lần chạy có số liệu so được với baseline đã chốt.

### ✅ Phase 0 (phần lõi) đã xong — a11y + TinyGo-default
- **0.2 a11y**: `Heading{Level}` → `<h1>`–`<h6>` thật (margin:0; spec giữ size/weight)
  → screen-reader + SEO thấy cấu trúc heading. `Link{Href}` → anchor thật (crawlable),
  rỗng thì fallback no-op JS link. `Image{Alt}` + `IconButton{Tooltip→aria-label}` đã
  có sẵn. Test: `TestHeadingRendersSemanticTag` (h1–h6 + margin:0), `TestLinkHref`.
  *Còn lại*: `Body`/`Caption`/`Badge` vẫn `<span>`, chưa có landmark roles — ghi docs.
- **0.1 TinyGo-default**: `gutter build deploy` mặc định dùng TinyGo nếu có trên PATH
  (`resolveDeployTinygo`); `--pure-go` opt-out, `--tinygo` ép, thiếu TinyGo thì fallback
  Go-std kèm gợi ý. `build`/`run` vẫn Go-std (compile nhanh). Cắt bundle ~60% cho prod.
- **0.3 perf gate CI**: hoãn (tác vụ CI/ops thuần; bench đã chạy tay được qua `bench/`).
- Toàn bộ test (host+wasm) + vet 2 target xanh; CLI build OK.

---

## Phase 1 — SSR: render cây widget ra HTML bằng Go thuần  ⭐ canh bạc lớn nhất

Mục tiêu: server trả HTML hoàn chỉnh → FCP/LCP tức thì + SEO, WASM tải nền. Đây là
đòn đảo chiều mục 1–3 của benchmark.

**Khả thi vì:** `Build()`/`Host()` đã platform-neutral và **compile được trên host**
(`app_stub.go` là bằng chứng pattern). SSR **không cần** Element tree (`element_wasm.go`),
chỉ cần đi đệ quy Widget → Host → HTML. Closure `OnMount` (chạm `syscall/js`) không
được gọi khi SSR nên không kéo `syscall/js` vào.

### 1.1 Core renderer  — **L**
- **File mới `ssr.go`** (KHÔNG build tag → compile cả host lẫn wasm; cấm import
  `syscall/js`):
  ```go
  package gutter

  // RenderToHTML walks a widget tree to HTML on the server. No DOM, no
  // syscall/js. State is built once (no SetState); OnMount/Events are ignored
  // but their presence is recorded as hydration markers.
  func RenderToHTML(root Widget, opts ...Option) (string, error)

  // RenderResult carries the HTML plus metadata hydration needs (e.g. which
  // node paths had event handlers / keys) — emitted as data-attrs.
  type RenderResult struct { HTML string; /* head hints, etc. */ }
  ```
- **Thuật toán** (gương của `newElement` type-switch tại `element_wasm.go:45`):
  - `HostWidget` → lấy `Host()`, ghi `<Tag attrs style>`, escape `Text`, đệ quy
    `Children`, đóng tag (xử lý void elements: `img,input,br,...`).
  - `StatelessWidget` → `Build(ctx)`, đệ quy kết quả.
  - `StatefulWidget` → `CreateState()`, gọi `InitState()` nếu có, `Build(ctx)`, đệ quy.
    (Server chỉ render trạng thái khởi tạo — không SetState, không Dispose.)
  - `Keyed` → phát `data-gutter-key="..."` để hydration khớp.
  - Đánh dấu node có `Events`/`OnMount` bằng `data-gutter-h="1"` (hydration biết
    chỗ nào cần "thức dậy").
- **Phụ trợ cần viết host-safe:** escape HTML attr/text; serialize `Style` map →
  `style="..."` (đã có logic ở `applyStyle`, tách phần thuần-chuỗi dùng chung).
- **BuildContext:** dùng lại cơ chế `ctx.Theme` hiện tại (Scaffold mutate ctx). Vì
  cùng một `*BuildContext` threaded xuống, SSR hoạt động y như client.
- **Rủi ro:** widget nào lỡ chạm `syscall/js` ngoài `*_wasm.go` sẽ vỡ host build —
  CI đã có `go build ./...` host nên bắt được sớm. Map `Events` chứa func không
  serialize được → chỉ ghi marker, không ghi handler (đúng thiết kế).
- **Test:** layer-1 — `RenderToHTML(Scaffold{...})` ra HTML mong đợi cho từng widget
  catalog (mở rộng `hostOf` thành `htmlOf`); golden-file cho counter/showcase.

### 1.2 SSR server + CLI  — **M**
- **`cmd/gutter`**: thêm `gutter ssr` / cờ `--ssr` cho `run`/`build deploy`:
  - Build 2 artifact: `app.wasm` (client, như cũ) **và** một binary host
    `ssr-server` (build thường) import app's `RootWidget()` để gọi `RenderToHTML`.
  - Cần app expose root cho cả hai entry: convention `func Root() gutter.Widget`
    dùng chung bởi `main` (wasm → `RunApp(Root())`) và ssr-server (→ `RenderToHTML(Root())`).
    Cập nhật `gutter new` scaffold theo pattern này.
  - Server: trả HTML từ `RenderToHTML`, nhúng `<div id=app>…ssr html…</div>` +
    script nạp `app.wasm`. Set `Cache-Control`, gzip (đã có mẫu ở `bench/server.mjs`).
- **Done khi:** `curl` ra HTML đầy đủ (xem được bằng JS tắt); Lighthouse SEO pass.

### 1.3 Đo lại để chứng minh  — **S**
- Thêm biến thể "gutter-ssr" vào `bench/run.mjs`; kỳ vọng FCP/LCP tụt về sát React,
  cold-ready không còn cõng instantiate trước first paint.
- **Done khi:** bảng mục 2 cho gutter-ssr ≈ React ở FCP/LCP.

### ✅ Trạng thái: PoC đã xong (1.1) — đã chứng minh
- `ssr.go` (`RenderToHTML`) + test layer-1 (`ssr_test.go`, `widgets/ssr_test.go`)
  xanh, build cả host lẫn wasm, không phá test cũ.
- Demo `bench/ssr-demo/` (Root() dùng chung wasm + host ssrgen) + `bench/run-ssr.mjs`.
- **Kết quả đo (localhost, median 5 cold load):**

  | Biến thể | FCP | LCP |
  |---|--:|--:|
  | CSR (WASM render client) | 181.8ms | 181.8ms |
  | **SSR (RenderToHTML)** | **37.5ms** | **37.5ms** |

  → SSR paint nội dung **nhanh 4.9×**. Trên mạng thật khoảng cách còn lớn hơn nhiều
  (CSR phải tải xong 2.8MB WASM *trước khi* vẽ bất cứ thứ gì; SSR vẽ ngay từ HTML).
### ✅ DX nâng cấp — `gutter.Serve(gutter.Config{...})` (một main cho cả 2 mode)
- `serve.go` (`Config{Root, RPC, Theme, Selector, Addr, Dist, Head}`) + `serve_wasm.go`
  (`Serve` = `RunApp(Root(), WithHydrate())`) + `serve_host.go` (`Serve` = đăng ký RPC +
  SSR + mount `/rpc` + static; `serveHandler` tách ra để test). Người dùng viết **một**
  `main.go`, **không build tag**, RPC đăng ký **một lần** ở `Config.RPC`.
- CLI `gutter run --ssr` giờ `go run .` (build host của chính main đó) thay vì cần
  `./server`. `gutter new` scaffold sẵn `gutter.Serve(gutter.Config{Root: Root})` →
  chạy được cả `gutter run` (CSR) lẫn `gutter run --ssr` không sửa gì.
- Test `serve_host_test.go`: một handler phục vụ cả SSR (`/`) lẫn RPC (`/rpc`). Smoke
  `gutter run --ssr` trên `examples/fullstack` (đã gộp về 1 `main.go`): SSR HTML +
  `{"sum":42}` + wasm đúng MIME. ✓

### ✅ Phase 1.2 đã xong — `ServeSSR` + CLI `--ssr`
- `ssr_server.go`: `gutter.ServeSSR(SSRConfig)` / `SSRHandler` (host-only) — render
  `Root()` mỗi request thành full HTML doc + bootstrap, phục vụ `app.wasm`/`wasm_exec.js`
  từ `Dist`. Test `ssr_server_test.go` (httptest) xanh.
- CLI `gutter run --ssr`: build wasm → `dist/`, rồi `exec go run ./server`
  (truyền `GUTTER_ADDR`/`GUTTER_DIST`); báo lỗi kèm hướng dẫn nếu thiếu `./server`.
  Convention: `app.Root()` dùng chung + `main_wasm.go` (`RunApp(..., WithHydrate())`)
  + `server/main.go` (`ServeSSR`). Reference: `bench/ssr-demo/`.
- **Smoke test CLI** (`gutter run --ssr` trên ssr-demo): trả full HTML có nội dung
  render (Dashboard/Feature 1/Likes: 0) + bootstrap, `app.wasm` đúng MIME `application/wasm`. ✓
- **Còn lại Phase 1:** 1.3 (gắn biến thể gutter-ssr vào `bench/run.mjs` để so cạnh
  React trong cùng bảng) — hiện đã có `bench/run-ssr.mjs` đo riêng. Và `gutter new --ssr`
  scaffold (hiện hướng dẫn copy `bench/ssr-demo/`).

---

## Phase 2 — Hydration: WASM "nhận" DOM của SSR thay vì dựng lại  (phụ thuộc Phase 1)

Mục tiêu: sau khi HTML SSR đã hiển thị, WASM gắn sự kiện/đời sống vào **đúng node
sẵn có** thay vì createElement + thay thế (tránh nhấp nháy & double-render).

### 2.1 Hydrate mount path  — **XL** (phần khó nhất của toàn roadmap)
- **`element_wasm.go`**: thêm chế độ hydrate vào pipeline mount:
  - `RunApp(root, WithHydrate())` → thay vì `mount` (createElement), gọi `hydrate`.
  - `hostElement.hydrate(existing js.Value, ctx)`: **không** createElement; nhận
    node SSR có sẵn, verify `tagName`/`data-gutter-key` khớp, gắn `syncEvents`,
    chạy `OnMount`, rồi `hydrate` đệ quy children theo thứ tự (dùng `childNodes`).
  - `statelessElement`/`statefulElement.hydrate` → build, hydrate child với cùng node.
  - **Mismatch policy:** nếu tag/cấu trúc lệch (SSR ≠ client build) → fallback
    `mount` (dựng lại subtree) + cảnh báo dev. An toàn trước hết.
- **Bỏ qua text-node fiddliness:** SSR phát text trùng cách client set `textContent`
  (đã thống nhất ở Phase 1) để khớp.
- **Rủi ro:** thứ tự children, whitespace text node, attribute Go set runtime
  (controlled inputs qua `setStringPropIfDifferent`) — phải hydrate value sau khi gắn.
  Đây là chỗ tốn công nhất; làm từng nhóm widget, có e2e chặn hồi quy.
- **Test:**
  - layer-2: dựng DOM bằng `RenderToHTML` (chạy host renderer trong test rồi
    `innerHTML`), gọi `hydrate`, assert **DOM node identity giữ nguyên** (không bị
    thay), event bắn đúng, controlled input giữ caret.
  - layer-3: testapp SSR → hydrate → click counter → vẫn chạy; không "nháy".
- **Done khi:** counter + showcase SSR-hydrate không nhấp nháy, mọi e2e cũ xanh.

### ✅ Phase 2 đã xong — hydration hoạt động end-to-end
- `WithHydrate()` (`options.go`) + `Element.hydrate(node, ctx)` cho cả 3 loại
  element (`element_wasm.go`): adopt node SSR sẵn có, re-apply attrs/style/text
  idempotent, strip marker `data-gutter-*`, gắn listener + `OnMount`, đệ quy theo
  `node.children`; lệch tag → remount subtree. Container rỗng → fallback mount
  (một `main()` chạy cả SSR lẫn CSR).
- `RunApp` chọn hydrate vs mount theo `cfg.hydrate` + có sẵn children.
- **Test layer-2** (`hydrate_wasm_test.go`, chạy qua wasmbrowsertest): giữ nguyên
  node identity, click sau hydrate chạm handler (`count:0`→`count:1`), tag-mismatch
  fallback. Xanh. Toàn bộ test cũ (host + wasm) không vỡ.
- **E2E demo** (`bench/ssr-demo` + `bench/run-ssr.mjs`): SSR FCP **38ms** vs CSR
  **186ms (4.9×)**, và sau boot nút "Likes" tương tác được (`Likes: 0`→`Likes: 1`).

---

## Phase 3 — InheritedWidget / Dependency Injection ambient  (song song, hỗ trợ 1–2)

Mục tiêu: lấp mảnh Flutter-parity còn thiếu (xem "No InheritedWidget" trong limits).
Theme/locale/router/RPC-client truyền ngầm thay vì truyền tay.

### 3.1 `InheritedWidget` + `ctx.DependOn[T]`  — **L**
- **`context.go`**: `BuildContext` mang thêm bảng inherited theo type:
  ```go
  type BuildContext struct {
      Theme *themes.Theme
      inherited map[reflect.Type]any // private
  }
  func DependOn[T any](ctx *BuildContext) (T, bool)
  ```
- **Widget mới `Provider[T]{ Value T; Child Widget }`** (StatelessWidget) đặt giá
  trị vào ctx khi build subtree, khôi phục khi ra (giống cách Scaffold set Theme).
- **Tái cấu trúc Theme** thành một inherited dependency chuẩn (giữ `ctx.Theme` như
  alias tương thích ngược).
- **Rủi ro:** `*BuildContext` hiện là 1 instance dùng chung toàn cây → push/pop
  phải đúng thứ tự DFS. Cần test kỹ với rebuild lệch nhánh.
- **Test:** layer-1 provide/depend; layer-2 rebuild subtree thấy giá trị mới.
- **Done khi:** `ObserverBuilder`/router/RPC-client lấy được qua `DependOn` thay vì
  truyền pointer tay.

### ✅ Phase 3 đã xong — Provider / DependOn
- `inherited.go` (neutral): `Provider[T]{Value, Child}` + `DependOn[T](ctx)` +
  interface ẩn `inheritedProvider`. `BuildContext.inherited map[reflect.Type]any`.
- Runtime (`element_wasm.go`): statelessElement push/pop scope ở mount/update/hydrate;
  **statefulElement capture `e.scope` lúc mount + refresh khi update, restore lúc
  `rebuild()`** → DependOn đúng cả khi **isolated SetState rebuild** (ca khó nhất mà
  stack đơn thuần sẽ sai). SSR (`ssr.go`) push/pop trong vòng đệ quy.
- Quyết định phạm vi: **chưa** route `Theme` qua DI (giữ field riêng, ít rủi ro);
  **chưa** có fine-grained invalidation (đổi `Value` lan theo top-down rebuild;
  giá trị đổi nhiều → dùng `Notifier` + `ObserverBuilder`). Đã ghi docs.
- **Test**: layer-1 (`inherited_test.go`) scope theo subtree + nested shadowing +
  absent/nil qua SSR; layer-2 (`inherited_wasm_test.go`) **DependOn sống sót qua
  isolated rebuild** (`live:0`→`live:1`, không rớt về `none`). App không dùng
  Provider ⇒ `inherited==nil`, zero overhead. Toàn bộ test cũ + vet 2 target xanh.

---

## Phase 4 — Typed Go full-stack RPC  ⭐ moat (độc lập, có thể chạy song song Phase 1–3)

Mục tiêu: gọi hàm server từ client bằng **struct Go dùng chung, không codegen,
không REST tay, không lệch type** — thứ React+Go không làm mượt được.

### 4.1 Gói `gutter/rpc`  — **L**
- **Định nghĩa một lần, dùng cả hai phía:**
  ```go
  // shared package
  type Greet struct{ Name string }            // request
  type Greeting struct{ Text string }         // response
  ```
- **Server (host build):** `rpc.Handle(func(ctx, Greet) (Greeting, error))` →
  tự đăng ký HTTP endpoint, (de)serialize JSON.
- **Client (wasm build, `*_wasm.go`):** `rpc.Call[Greet, Greeting](req)` →
  `fetch` qua `syscall/js`, parse về struct. `*_stub.go` cho host.
- **An toàn type bằng generics** ở cả hai đầu, cùng struct → đổi field là compiler
  bắt lỗi trên toàn stack (điểm bán hàng cốt lõi).
- **Tích hợp:** `AsyncBuilder` + `rpc.Call` = data-fetching khai báo; client lấy
  từ `DependOn` (Phase 3).
- **Bảo mật:** document rõ validate/authz phía server (giống cảnh báo ở
  `community/login_with_google`).
- **Test:** layer-1 (de)serialize round-trip; e2e: testapp gọi RPC tới chính
  ssr-server, hiển thị kết quả.
- **Done khi:** ví dụ `examples/fullstack` chạy 1 lệnh, share struct, đổi field thì
  build fail đúng chỗ.

### ✅ Phase 4 đã xong — typed RPC hoạt động end-to-end
- Gói `rpc/` (chỉ import stdlib): `Handle[Req,Res]` + `Handler()` (server) và
  `Call[Req,Res]` (client). Route suy ra từ **kiểu Req** (`reflect.TypeFor`) → client/
  server tự khớp, không codegen, không chuỗi. Client dùng `net/http` (wasm → Fetch).
- Test `rpc/rpc_test.go` (httptest, host): round-trip (`2+40=42`), propagate lỗi
  handler, unknown-proc 404, duplicate-handler panic. Xanh. Build cả host+wasm.
- Ví dụ `examples/fullstack/`: `api` (struct dùng chung) + `app.Root()` (UI gọi RPC)
  + `main_wasm.go` (`RunApp(..., WithHydrate())`) + `server/main.go` (RPC `/rpc` +
  SSR `/` trong một process). **Smoke test**: `POST /rpc` trả `{"sum":42}`, unknown→404,
  `/` trả SSR HTML. ✓ Đổi field trong `api` ⇒ vỡ compile cả hai phía.
- **Lưu ý:** `net/http` làm `app.wasm` phình (~11MB raw); cân nhắc TinyGo/gzip, hoặc
  sau này một transport fetch gọn hơn cho client.

---

## Phase 5 — Island / embed mode  (phụ thuộc 1–2)

Mục tiêu: nhúng 1 widget Gutter vào trang HTML/React có sẵn; **chỉ nạp WASM cho
island** → né hẳn bài toán bundle, cho adoption từng phần.

### 5.1 Multi-root mount + lazy boot  — **L**
- `RunApp` hiện mount 1 selector. Thêm `MountInto(selector, widget)` cho **nhiều
  island** trong một trang, mỗi island hydrate vùng SSR của nó.
- **Lazy:** chỉ `instantiateStreaming` khi island vào viewport (IntersectionObserver,
  `*_wasm.go`) → trang tĩnh hiện ngay, WASM chỉ trả giá ở island tương tác.
- **Đóng gói:** CLI emit snippet nhúng (`<div data-gutter-island="...">` + loader).
- **Test:** e2e trang HTML thường + 2 island; chỉ island được hydrate, phần còn lại
  không cần WASM.
- **Done khi:** demo "1 island Gutter trong trang HTML tĩnh" tải nhanh như trang tĩnh.

### ✅ Phase 5 đã xong — islands / multi-root
- `MountInto(selector, root, opts...) *App` (non-blocking, hỗ trợ `WithHydrate`) +
  `MountWhenVisible` (IntersectionObserver, mount khi vào viewport). Refactor `RunApp`
  = `MountInto(cfg.selector, …)` + `select{}`. Stub host trong `app_stub.go`.
- Test layer-2 (`island_wasm_test.go`): 2 island độc lập (click A không ảnh hưởng B),
  hydrate island. Xanh.
- Ví dụ `examples/islands/`: trang HTML tĩnh + 2 island + **lazy loader** (IntersectionObserver
  fetch `app.wasm` lần đầu island vào tầm nhìn). Smoke (`bench/run-islands.mjs`):
  island 1 tương tác (`(0)`→`(1)`), island 2 độc lập. ✓
- Ý nghĩa: né bài toán bundle — trang tĩnh trả **0 WASM** tới khi cần tương tác.

---

## Phase 6 — Hardening để ship thật  (xuyên suốt)

- **Forms + validation** (**M**): widget `Form`/`FormField` + validate dùng chung
  struct với RPC (Phase 4).
- **Router trưởng thành** (**M**): guards, nested routes, query parsing (xem limits).
- **DevTools** (**M**): overlay cây Element + đếm rebuild (build dev), từ `flushRebuilds`.
- **Ecosystem** (**ongoing**): docs site (chính SSR), templates `gutter new --template`,
  mở rộng catalog `community/`.

### ✅ Phase 6 (phần lõi) đã xong — Forms + Router query
- **Forms + validation**: `widgets/validation.go` — `Validator func(string) string`
  thuần + built-ins `Required`/`MinLength`/`MaxLength`/`Email`/`Pattern`/`Combine`
  (chia sẻ được với check phía server cạnh `gutter/rpc`). `widgets/form.go` —
  `Form{Fields, Submit, OnSubmit}` + `FormField`: controlled, hiện lỗi inline,
  `OnSubmit` chỉ chạy khi mọi field hợp lệ. Test layer-1: validators, `Combine`,
  `validateFields`, `Form.InitState`+`Build`.
- **Router query**: `Router.Query() url.Values` + match **strip query** (
  `/user/42?tab=x` vẫn khớp `/user/:id`). Test layer-1: strip-query-khi-match,
  parse query (`q=go+lang`→"go lang"), path không query → rỗng.
- Toàn bộ test (host+wasm) + vet 2 target xanh.
- **Hoãn** (polish, ít rủi ro, làm sau): router guards/nested/transitions,
  DevTools overlay, docs site, `gutter new --template`.

---

## Sơ đồ phụ thuộc & thứ tự đề xuất

```
Phase 0 (quick wins) ─┐
                      ├─> Phase 1 SSR ─> Phase 2 Hydration ─> Phase 5 Islands
Phase 3 DI  ──────────┘                                  ┌─> Phase 6 Hardening
Phase 4 RPC (song song, độc lập) ────────────────────────┘
```

**Milestone gợi ý:**
- **M1 "Nhẹ cửa hơn"** = Phase 0. (Cắt bundle/CPU ngay, có a11y, có gate đo.)
- **M2 "Mở SEO/landing"** = Phase 1 + 2. (Đảo chiều cold-start — quan trọng nhất.)
- **M3 "Moat full-stack"** = Phase 3 + 4. ("Viết cả web bằng Go, type-safe.")
- **M4 "Adoption"** = Phase 5 + 6. (Nhúng dần, đủ chín để ship.)

**Tiêu chí thành công tổng:** sau M2, bảng `bench/RESULTS.md` cho biến thể SSR có
FCP/LCP ≈ React; sau M3, có ví dụ full-stack một-ngôn-ngữ chạy được mà React-stack
không sánh được về type-safety. Lúc đó "use-case hẹp" trở thành "niche sâu, đủ sống".
