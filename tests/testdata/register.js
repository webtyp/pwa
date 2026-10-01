if ("serviceWorker" in navigator) {
  const sw = navigator.serviceWorker;
  const hadController = !!sw.controller;
  let reloading = false;
  sw.addEventListener("controllerchange", () => {
    if (!hadController || reloading) return;
    reloading = true;
    location.reload();
  });
  sw.register("/sw.js").then((reg) => {
    const notify = () => window.dispatchEvent(new Event("webtyp-update-ready"));
    const watch = (w) => w && w.addEventListener("statechange", () => {
      if (w.state === "installed" && sw.controller) notify();
    });
    if (reg.waiting && sw.controller) notify();
    watch(reg.installing);
    reg.addEventListener("updatefound", () => watch(reg.installing));
  });
}