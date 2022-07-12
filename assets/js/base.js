function setThemeMeta(color) {
  let meta = document.querySelector("meta[name='theme-color']");
  meta.setAttribute("content", color);
}

function setTheme(theme) {
  localStorage.setItem("theme", theme);

  let current = document.getElementById("current-theme");
  if (current) {
    current.remove();
  }

  let head = document.getElementsByTagName("head")[0];
  let link = document.createElement("link");
  link.rel = "stylesheet";
  link.id = "current-theme";
  link.type = "text/css";
  link.href = theme;
  head.appendChild(link);

  setTimeout(() => {
    let color = getComputedStyle(document.documentElement).getPropertyValue(
      "--mantle"
    );
    localStorage.setItem("theme-base", color);
    setThemeMeta(color);
    let el = document.getElementById("theme-select");
    el.value = theme;
  }, 33);
}

window.setTheme = setTheme;
window.setThemeMeta = setTheme;

let theme = localStorage.getItem("theme") || "/static/css/themes/macchiato.css";
let color = localStorage.getItem("theme-base") || "#1e2030;";

setTheme(theme);
setThemeMeta(color);

// set theme
window.onload = () => {
  let selector = document.getElementById("theme-select");
  selector.value = theme;
};

// good scroll restore
if ("scrollRestoration" in history) {
  history.scrollRestoration = "auto";
}
