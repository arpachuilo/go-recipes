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

    let isLightMode = theme === "/css/themes/light.css";
    console.log(isLightMode, document.getElementById("light"))
    document.getElementById("light").checked = isLightMode;
    document.getElementById("dark").checked = !isLightMode;
  }, 33);
}

window.setTheme = setTheme;
window.setThemeMeta = setTheme;

let theme = localStorage.getItem("theme") || "/static/css/themes/dark.css";
let color = localStorage.getItem("theme-base") || "##141414";

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

// service worker
const registerServiceWorker = async () => {
  if ("serviceWorker" in navigator) {
    try {
      const registration = await navigator.serviceWorker.register(
        "/service-worker.js",
        {
          scope: "/",
        }
      );

      if (registration.installing) {
        console.log("Service worker installing");
      } else if (registration.waiting) {
        console.log("Service worker installed");
      } else if (registration.active) {
        console.log("Service worker active");
      }

      // After the initial load, force a service worker update check each time
      // our web app is hidden and then brought back to the foreground.
      document.addEventListener("visibilitychange", () => {
        if (document.visibilityState === "visible") {
          registration.update();
        }
      });
    } catch (error) {
      console.error(`Registration failed with ${error}`);
    }
  }
};

registerServiceWorker();
