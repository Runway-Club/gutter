# Gutter vs React — phân tích hiệu năng render & reload

So sánh hai app **giống hệt nhau về cấu trúc** ở 4 mức độ phức tạp (số item =
10 / 100 / 1.000 / 10.000). Mỗi item là một `<button>` có state đếm riêng. Bản
Gutter biên dịch ra WASM (Go std và TinyGo), bản React build production bằng
Vite. Đo bằng Playwright/Chromium, server tĩnh có gzip + cache.

Raw data: `RESULTS.md` / `results.json`. Cách chạy lại ở cuối file.

## 1. Bundle size — thứ trình duyệt phải tải

| Framework | Raw | Gzip |
|---|--:|--:|
| Gutter (Go WASM) | 2764 KB | **792 KB** |
| Gutter (TinyGo WASM) | 943 KB | **306 KB** |
| React (Vite prod) | 140 KB | **45 KB** |

Bundle **cố định**, không tăng theo số item. Đây là chi phí lớn nhất và nó
nghiêng hẳn về React: bundle Gutter (Go) nặng **~17,5×** React; TinyGo thu hẹp
còn ~6,8× nhưng vẫn nặng hơn nhiều.

Vì benchmark chạy trên localhost (tải gần như tức thì), con số "cold" bên dưới
**chưa tính** thời gian tải bundle thật. Ước tính thời gian tải gzip theo băng thông:

| | React 45 KB | TinyGo 306 KB | Go 792 KB |
|---|--:|--:|--:|
| Broadband 50 Mbps | ~7 ms | ~48 ms | ~124 ms |
| 4G ~10 Mbps | ~35 ms | ~240 ms | ~620 ms |
| 3G/yếu ~1.6 Mbps | ~220 ms | ~1.5 s | ~3.9 s |

→ Trên mạng thật, phải **cộng thêm** các mức này vào cold render. Khoảng cách cold
giữa Gutter và React ngoài đời lớn hơn nhiều so với số đo localhost.

## 2. Cold render (mạng trống, cache rỗng) — ms

| Items | Gutter Go | Gutter TinyGo | React |
|--:|--:|--:|--:|
| 10 | 138.9 | 54.3 | **19.5** |
| 100 | 146.3 | 50.1 | **19.7** |
| 1.000 | 167.0 | 62.7 | **26.9** |
| 10.000 | 315.4 | 198.3 | **86.6** |

- **Chi phí cố định lúc khởi động chi phối ở app nhỏ.** Ở n=10, Gutter (Go) đã
  tốn ~140 ms chỉ để instantiate + chạy runtime WASM trước khi vẽ — gấp ~7× React.
  Đây là phần app càng đơn giản càng "đau": không có gì để khấu hao.
- **Càng phức tạp, khoảng cách *tương đối* càng hẹp.** n=10 → Gutter Go ≈ 7× React;
  n=10.000 → chỉ còn ~3,6×. Vì chi phí khởi động cố định bị pha loãng bởi chi phí
  render tăng dần (cái mà cả hai đều phải trả).
- **TinyGo cắt ~⅔ chi phí khởi động** (54 ms vs 139 ms ở n=10), kéo Gutter lại gần
  React đáng kể ở các tier nhỏ.

FCP và LCP gần như trùng nhau và bám sát cold-render (cả lưới vẽ trong một nhịp paint).

## 3. Warm reload (cache đã nóng) — ms

| Items | Gutter Go | Gutter TinyGo | React |
|--:|--:|--:|--:|
| 10 | 18.7 | 18.8 | **15.5** |
| 100 | 18.0 | 18.5 | **17.7** |
| 1.000 | 41.4 | 34.9 | **25.4** |
| 10.000 | 196.3 | 171.9 | **96.4** |

- **Khi loại bỏ tải mạng, khoảng cách sụp đổ.** Ở app nhỏ/vừa (≤100 item) cả ba
  gần như ngang nhau (~16–19 ms) — instantiate WASM từ cache + render rất nhanh.
- Go và TinyGo **gần như bằng nhau** khi warm: lúc này chênh lệch bundle không còn
  ý nghĩa, chỉ còn chi phí compile/instantiate đã được cache phần lớn.
- Ở 10.000 item, render thắng thế: Gutter ~2× React. Đây là chi phí dựng cây
  10.000 stateful element phía WASM so với 10.000 component React.

## 4. Update latency (click → DOM cập nhật, median) — ms

| Items | Gutter Go | Gutter TinyGo | React |
|--:|--:|--:|--:|
| 10 | 2.0 | 2.4 | 2.6 |
| 100 | 2.3 | 2.1 | 2.4 |
| 1.000 | 3.0 | 2.5 | 2.8 |
| 10.000 | 3.1 | 3.2 | 3.2 |

- **Hòa.** Cập nhật một item trong cây lớn tốn ~2–3 ms ở cả ba, gần như không tăng
  theo kích thước cây. Cả hai reconciler đều chỉ động vào đúng subtree thay đổi:
  `SetState` batch theo microtask của Gutter ngang ngửa re-render component của React.
- Đây là điểm Gutter **không** thua: một khi đã chạy, chi phí cập nhật runtime là
  tương đương. Lợi thế của React nằm hoàn toàn ở khởi động & kích thước tải.

## 5. Kết luận theo độ phức tạp tăng dần

| | App rất nhỏ (~10) | Vừa (~100–1.000) | Lớn (~10.000) |
|---|---|---|---|
| **Cold render** | React thắng đậm (~7×). Gutter trả phí khởi động cố định mà không khấu hao được | React vẫn dẫn ~6×, nhưng app vẫn "đủ nhanh" với cả hai | React dẫn ~3,6×; khoảng cách tương đối hẹp nhất nhưng chênh tuyệt đối lớn nhất |
| **Warm reload** | Gần ngang (~3–4 ms chênh) | Gần ngang | React ~2× nhanh hơn |
| **Update** | Hòa | Hòa | Hòa |
| **Bundle** | React 45 KB vs Gutter 792 KB — quyết định trải nghiệm lần đầu | như nhau (cố định) | như nhau (cố định) |

**Tóm lại:**

1. **React vượt trội ở lần tải đầu (cold), nhất là app nhỏ.** Gutter gánh chi phí
   cố định: bundle WASM lớn (792 KB gz, ~17× React) + ~100–140 ms instantiate runtime.
   App càng đơn giản, phí cố định này càng vô lý vì không có gì để khấu hao.
2. **Độ phức tạp tăng làm hẹp khoảng cách *tương đối*** (7× → 3,6×) vì chi phí render
   theo quy mô cây là thứ cả hai cùng trả; nhưng Gutter **không** lật ngược được — nó
   chỉ đuổi gần lại chứ không vượt.
3. **Sau lần đầu (warm) và khi tương tác, hai bên gần như ngang nhau.** Reconciler
   của Gutter cạnh tranh sòng phẳng; điểm yếu duy nhất là khởi động + tải.
4. **TinyGo là đòn bẩy lớn nhất cho Gutter**: cắt bundle còn 306 KB và khởi động còn
   ~50 ms, xóa phần lớn bất lợi ở tier nhỏ — đáng cân nhắc cho production.

**Khi nào Gutter hợp lý:** app sống lâu trong tab (SPA, dashboard nội bộ, app sau
đăng nhập) — nơi cold load chỉ trả một lần còn lợi ích "viết UI bằng Go" được hưởng
mãi. **Khi nào React hợp lý hơn:** landing/marketing, app cần tải lần đầu nhanh trên
mobile/3G, hoặc app rất nhỏ — nơi 792 KB WASM là cái giá quá đắt.

## 6. Memory footprint & CPU

Đo bằng CDP (`run-resource.mjs`), median 2 lần chạy. Raw: `RESOURCES.md`.
**Memory** = JS heap + WASM linear memory (heap của Go nằm trong linear memory,
**không** nằm trong JS heap — nên `performance.memory` đơn thuần sẽ bỏ sót gần hết
bộ nhớ của Gutter). **CPU** = thời gian main-thread bận (`TaskDuration`).

### Memory total (MB) — JS heap + WASM

| Items | Gutter Go | Gutter TinyGo | React |
|--:|--:|--:|--:|
| 10 | 4.3 | **1.5** | **1.5** |
| 100 | 5.3 | 1.9 | **1.6** |
| 1.000 | 8.0 | 4.4 | **2.9** |
| 10.000 | 34.6 | 52.0 | **14.9** |

- **React tiết kiệm RAM nhất ở mọi tier** (trừ tier nhỏ thì TinyGo hòa). Heap của
  Gutter Go nằm trong WASM linear memory: ~3–6 MB cố định + tăng theo cây.
- **TinyGo nhỏ nhất ở app nhỏ/vừa nhưng phình ở quy mô lớn**: ở 10.000 item nó ngốn
  **52 MB** — *nhiều hơn cả Go std (34,6 MB)*. GC của TinyGo cấp phát rộng tay và
  thu hồi kém khi allocation dày. ⚠️ TinyGo chỉ thắng RAM ở app nhỏ–vừa.
- JS heap của Gutter rất nhỏ (1,3–3,6 MB) vì object Go sống trong linear memory.
  JS heap của React lại là toàn bộ footprint (14,9 MB ở 10.000) → React không có
  "phần ẩn" như WASM.

### CPU cold render — main-thread bận (ms)

| Items | Gutter Go | Gutter TinyGo | React |
|--:|--:|--:|--:|
| 10 | 23.9 | 14.2 | **13.2** |
| 100 | 28.4 | 16.7 | **14.7** |
| 1.000 | 47.3 | 33.7 | **25.0** |
| 10.000 | 241.2 | 203.4 | **91.9** |

- **React tốn ít CPU nhất.** Gutter cõng thêm thuế cố định ~15–20 ms để
  compile + instantiate module WASM (ở n=10: tổng 23,9 ms nhưng layout chỉ 2,1 ms
  và "script" 0,6 ms → ~21 ms còn lại là compile/instantiate/init Go runtime).
  ⚠️ Chromium **không tính** thời gian chạy WASM vào `ScriptDuration` — nên với
  Gutter phải đọc cột "main-thread bận", đừng đọc cột script.
- Ở 10.000 item, dựng cây thắng thế: Go ~2,6× CPU của React.

### CPU mỗi lần update (µs: script + layout)

| Items | Gutter Go | Gutter TinyGo | React |
|--:|--:|--:|--:|
| 10 | 270+130 | 568+120 | 182+137 |
| 100 | 284+208 | 516+166 | 200+216 |
| 1.000 | 170+575 | 390+557 | 206+532 |
| 10.000 | 103+2510 | 108+2535 | 539+2857 |

- **Ở quy mô lớn, layout/reflow của trình duyệt chi phối** (~2,5–2,9 ms/update ở
  10.000) và nó **độc lập framework** — đổi text 1 nút trong container flex 10.000
  con buộc trình duyệt reflow lớn. Vì thế update latency hội tụ (~3 ms, khớp mục 4).
- Phần script/update của Gutter Go nhỏ (~0,1–0,3 ms); **TinyGo cao hơn ~2×** ở tier
  nhỏ (mã sinh chậm hơn). Reconciler theo subtree của Gutter cạnh tranh tốt.

### DOM nodes

Gutter phát **ít node hơn** (1 text node/item) so với React (~6/item) — vì JSX tách
mỗi biểu thức `{...}` kề nhau thành text node riêng. Đây là khác biệt do *cách viết*
(idiom), không phải bản chất framework; nó làm React nhiều node + layout nặng hơn,
nhưng React vẫn thắng tổng RAM.

### Chốt memory/CPU

- **RAM:** React < TinyGo (app nhỏ) ≪ Go (app lớn); TinyGo phản tác dụng ở app rất lớn.
- **CPU render:** React thấp nhất; Gutter trả thuế WASM cố định + chi phí dựng cây.
- **CPU update:** gần như hòa, vì layout của trình duyệt mới là phần đắt nhất.
- Footprint của Gutter "ẩn" trong WASM linear memory — công cụ đo JS-heap thuần sẽ
  báo sai (thấp giả tạo).

## 7. Compute nặng trong browser (Go-WASM vs JS vs TinyGo)

3 kernel **viết tay y hệt** ở Go và JS (không stdlib sort, không Web Crypto), dữ
liệu sinh trong hàm bằng cùng PRNG, warmup cho JIT, median 5 lần. **Checksum khớp
9/9** → xác nhận cả ba làm đúng cùng một việc. Times = ms, thấp hơn = tốt hơn.
"speedup" = thời gian JS ÷ Go (>1 nghĩa là Go-WASM nhanh hơn).

| Kernel (size lớn nhất) | JS (V8) | Go-WASM | TinyGo | Go vs JS |
|---|--:|--:|--:|--:|
| mandelbrot 1024² (float thuần) | 93 | **83** | 89 | **1.12×** |
| quicksort 4M (int, nhánh, bộ nhớ) | **252** | 334 | 232 | 0.76× |
| matmul 384³ (FLOPs) | 36 | 48 | **20** | 0.75× |

**Kết quả thành thật (không tô hồng):**

- **V8 quá giỏi.** Không có chuyện "WASM đè bẹp JS". Với các vòng lặp số học đơn
  giản, JIT của V8 sinh mã gần như tối ưu. Go-std-WASM chỉ **ngang ngửa, lệch
  ±10–30%**.
- **Go-std thắng ở float loop thuần** (mandelbrot, +12%) nhưng **thua ở việc nhiều
  nhánh/truy cập mảng** (sort, matmul −25%) — vì backend WASM của Go std chèn nhiều
  bounds-check và tối ưu kém hơn.
- **Yếu tố quyết định không phải "WASM vs JS" mà là TRÌNH BIÊN DỊCH.** TinyGo (dùng
  LLVM) **nhanh hơn cả JS lẫn Go-std**: matmul **20 ms vs JS 36 ms (~1.8×)**,
  sort cũng nhỉnh hơn JS. Cùng là Go-ra-WASM nhưng TinyGo bỏ xa Go std ở compute số.

**Vậy "compute trong browser" có phải lý do chọn Gutter không?** Trung thực:

1. **Không phải vì nhanh gấp bội** — V8 quá mạnh, Go-std chỉ huề. Đừng bán Gutter
   bằng lời hứa "nhanh hơn JS".
2. **Nhanh hơn *có điều kiện*: dùng TinyGo cho kernel số học** (FLOP-heavy) thì
   thắng JS rõ (~1.8×).
3. **Lý do thật sự đáng giá:** (a) **tái dùng thư viện Go phức tạp** chạy đúng trong
   browser (parser, business logic, crypto, decoder…) thay vì viết lại bằng JS;
   (b) **hiệu năng ổn định, không có vách warmup/deopt** của JIT — quan trọng cho
   tác vụ dài, đều; (c) một ngôn ngữ cho cả stack.

→ Đây mới là bức tranh đúng: Gutter/Go-WASM **không thắng JS về tốc độ thô**, nhưng
**huề (và thắng nếu dùng TinyGo cho số học)**, cộng với lợi ích lớn về tái dùng code
và sự nhất quán ngôn ngữ. "Thua" ở mục 1–6 chỉ là thuế khởi động UI; ở compute thì
khoảng cách biến mất.

## 8. SSR vs React — head-to-head (sau khi triển khai SSR + hydration)

Cùng app grid như mục 2, thêm biến thể **Gutter SSR** (pre-render HTML + hydrate),
bản Go và TinyGo. Đây là cú trả lời trực tiếp cho mục 2 (lúc đó Gutter CSR thua
React 3–7×). Đo localhost, median 5 cold load.

### Cold render-complete (ms) — nav → đủ N item trong DOM

| Items | Gutter CSR | Gutter SSR (Go) | Gutter SSR (TinyGo) | React |
|--:|--:|--:|--:|--:|
| 10 | 249 | **26** | 28 | 43 |
| 100 | 200 | **31** | 36 | 40 |
| 1.000 | 222 | **37** | 41 | 43 |
| 10.000 | 376 | 76 | 81 | **63** |

### First Contentful Paint (ms) — SSR thắng React ở **mọi** tier

| Items | Gutter SSR (Go) | Gutter SSR (TinyGo) | React |
|--:|--:|--:|--:|
| 10 | **29** | 31 | 48 |
| 100 | **36** | 41 | 47 |
| 1.000 | **44** | 48 | 59 |
| 10.000 | **86** | 90 | 125 |

### Warm reload (ms)

| Items | Gutter SSR (Go) | Gutter SSR (TinyGo) | React |
|--:|--:|--:|--:|
| 10 | 11.5 | **6.9** | 14.9 |
| 100 | 17.2 | **8.0** | 15.1 |
| 1.000 | 14.6 | **12.3** | 25.0 |
| 10.000 | 70.6 | 74.8 | **96.1** → SSR thắng |

### Bundle (gzip)

| | React | Gutter TinyGo | Gutter Go |
|---|--:|--:|--:|
| gzip | **45 KB** | 407 KB | 955 KB |

Update latency: hòa ~2–3ms cả bốn (như mọi mục trước).

### Kết luận

1. **SSR xóa khoảng cách cold-start.** Mục 2: Gutter CSR thua React 3–7×. Giờ:
   **Gutter SSR thắng FCP ở mọi tier** (HTML paint ngay, không chờ chạy framework),
   cold-ready ngang/hơn React tới 1.000 item, chỉ thua ở 10.000 (parse khối HTML 977KB).
   **Warm reload SSR thắng toàn bộ.**
2. **Go-SSR ≈ TinyGo-SSR ở FCP/cold-ready** — vì các mốc này bị chi phối bởi
   **parse HTML**, không phải kích thước WASM (WASM tải *sau* first paint để hydrate).
3. **Lợi thế của TinyGo-SSR nằm ở chỗ khác**: bundle hydration **407KB vs 955KB gz**
   (→ time-to-interactive nhanh hơn nhiều trên mạng thật) và warm reload nhỉnh hơn
   (instantiate nhanh hơn). ⇒ **Production nên SSR + TinyGo.**
4. **Caveat**: localhost ⇒ FCP chưa tính tải mạng. Trên mạng thật, app nhỏ SSR HTML
   chỉ 1–9KB (paint tức thì) còn WASM (407KB–955KB gz) defer cho hydration — nên
   *time-to-content* thắng React, nhưng *time-to-interactive* vẫn gánh tải WASM.
   Bundle vẫn lớn hơn React (45KB) nhiều.

## Lưu ý về phương pháp (đọc số cho đúng)

- Mỗi item là **widget tối giản** (1 nút). App "sản phẩm" thật dùng widget themed
  giàu hơn → Gutter tốn thêm chi phí build cây, React tốn thêm JS lib; con số tuyệt
  đối sẽ cao hơn nhưng **xu hướng** không đổi.
- Đo trên localhost ⇒ cold render **chưa gồm** tải bundle thật (xem mục 1).
- "Render-complete" = thời điểm đủ N item có trong DOM (mốc tự đặt), không phải TTI
  chuẩn. Đo y hệt nhau cho cả hai nên so sánh là công bằng.
- Số đo trên một máy, một lần chạy; coi là **bậc độ lớn**, không phải con số tuyệt đối.

## Chạy lại

```sh
# build 3 bundle
cd bench/react-app && npm i && npm run build
cd ../gutter-app && /path/to/gutter build --tinygo && mv dist dist-tinygo && /path/to/gutter build
# chạy benchmark (mọi tier, hoặc truyền danh sách: node run.mjs 10,100)
cd ../ && npm i && npx playwright install chromium
node run.mjs            # render/reload  -> RESULTS.md
node run-resource.mjs   # memory + CPU   -> RESOURCES.md
```
