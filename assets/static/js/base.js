(()=>{function c(e){document.querySelector("meta[name='theme-color']").setAttribute("content",e)}function r(e){localStorage.setItem("theme",e);let o=document.getElementById("current-theme");o&&o.remove();let i=document.getElementsByTagName("head")[0],t=document.createElement("link");t.rel="stylesheet",t.id="current-theme",t.type="text/css",t.href=e,i.appendChild(t),setTimeout(()=>{let l=getComputedStyle(document.documentElement).getPropertyValue("--mantle");localStorage.setItem("theme-base",l),c(l);let s=document.getElementById("theme-select");s.value=e},33)}window.setTheme=r;window.setThemeMeta=r;var n=localStorage.getItem("theme")||"/static/css/themes/macchiato.css",a=localStorage.getItem("theme-base")||"#1e2030;";r(n);c(a);window.onload=()=>{let e=document.getElementById("theme-select");e.value=n};"scrollRestoration"in history&&(history.scrollRestoration="auto");var m=async()=>{if("serviceWorker"in navigator)try{let e=await navigator.serviceWorker.register("/service-worker.js",{scope:"/"});e.installing?console.log("Service worker installing"):e.waiting?console.log("Service worker installed"):e.active&&console.log("Service worker active")}catch(e){console.error(`Registration failed with ${e}`)}};m();})();
//# sourceMappingURL=base.js.map
