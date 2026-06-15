export function attachDragDrop(node, onPaths) {
  if (!node) return;
  node.style.setProperty("--wails-drop-target", "drop");

  const emitPaths = (paths) => {
    const realPaths = paths.filter((path) => typeof path === "string" && path.trim() !== "");
    if (realPaths.length > 0) {
      onPaths(realPaths);
    }
  };

  if (window.runtime?.OnFileDrop) {
    window.runtime.OnFileDrop((x, y, paths) => {
      node.classList.remove("dragging");
      if (!pointInNode(node, x, y)) return;
      emitPaths(paths || []);
    }, true);
  }

  node.addEventListener("dragover", (event) => {
    event.preventDefault();
    node.classList.add("dragging");
  });
  node.addEventListener("dragleave", () => node.classList.remove("dragging"));
  node.addEventListener("drop", (event) => {
    event.preventDefault();
    node.classList.remove("dragging");
    const paths = Array.from(event.dataTransfer?.files || [])
      .map((file) => file.path)
      .filter(Boolean);
    emitPaths(paths);
  });
}

function pointInNode(node, x, y) {
  const rect = node.getBoundingClientRect();
  return x >= rect.left && x <= rect.right && y >= rect.top && y <= rect.bottom;
}
