// strike out text
function strike(el) {
  el.classList.toggle("strike");
}

window.strike = strike;

// focus search
function slashToFocus(el) {
  window.addEventListener("keydown", (e) => {
    if (document.activeElement === el) return;
    if (e.code === "Slash") {
      el.focus();
      e.preventDefault();
    }
  });
}

window.slashToFocus = slashToFocus;

// scroll sync
function syncSrolls(_tx, _rx) {
  const tx = document.querySelector(_tx);
  const rx = document.querySelector(_rx);

  let hovered = false;
  let offset = 0;
  let scrollOfset = 0;
  const sync = () => {
    // only apply if below certain width
    if (window.innerWidth < 480) {
      offset = 0;
      scrollOfset = 0;
      return;
    }

    // compute max
    let max = tx.offsetHeight - rx.offsetHeight;
    max = Math.max(0, max);

    // sticky scroll
    let rxb = rx.getBoundingClientRect();
    if (hovered) {
      scrollOfset = Math.min(rxb.top, 0);
    }

    // reset sticky scroll
    if (rxb.top > 0) {
      scrollOfset = 0;
    }

    // no need to move
    if (offset < 0 && rx.style.marginTop == "0px") {
      return;
    }

    offset = tx.getBoundingClientRect().top * -1 + scrollOfset;
    offset = Math.max(0, offset);
    offset = Math.min(offset, max);
    rx.style.marginTop = offset + "px";
  };

  rx.addEventListener("mouseenter", (_) => {
    hovered = true;
  });

  rx.addEventListener("mouseleave", (_) => {
    hovered = false;
  });

  window.addEventListener("scroll", (_) => {
    window.requestAnimationFrame(sync);
  });
}

window.syncSrolls = syncSrolls;

// wake-lock for chrome
if ("wakeLock" in navigator) {
  navigator.wakeLock.request('screen');
}
