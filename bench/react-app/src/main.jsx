import React, { useState } from 'react';
import { createRoot } from 'react-dom/client';

// One stateful button per item — mirror of Gutter's benchItem. Clicking
// re-renders only this component (React reconciles its single <button>).
function Item({ i }) {
  const [c, setC] = useState(0);
  return (
    <button data-bench-item={i} style={{ padding: '4px 8px' }} onClick={() => setC(c + 1)}>
      Item {i} {'—'} {c}
    </button>
  );
}

function App({ n }) {
  const items = [];
  for (let i = 0; i < n; i++) items.push(<Item key={i} i={i} />);
  return (
    <div id="grid" style={{ display: 'flex', flexWrap: 'wrap', gap: '4px', padding: '8px' }}>
      {items}
    </div>
  );
}

const n = parseInt(new URLSearchParams(location.search).get('n') || '100', 10);
createRoot(document.getElementById('root')).render(<App n={n} />);
