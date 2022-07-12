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

// auto search limit size
function setSearchLimit() {
  let el = document.getElementById("search-results");
  if (el === null) return;
  let name = "search_limit";
  let value = Math.max(Math.floor(Math.min(1320, el.offsetWidth) / 200), 2) * 5;
  let expiry = new Date();
  let domain = window.location.hostname;
  expiry.setMonth(expiry.getMonth() + 2);
  let cookie = `${name}=${value};expires=${expiry};domain=${domain};path=/`;
  document.cookie = cookie;
}

window.onload = function () {
  if (!navigator.cookieEnabled) return;
  const urlSearchParams = new URLSearchParams(window.location.search);
  const params = Object.fromEntries(urlSearchParams.entries());
  let el = document.getElementById("limit");
  if ("limit" in params) {
    el.value = params.limit;
  } else {
    if (document.cookie.match(/^(.*;)?\s*search_limit\s*=\s*[^;]+(.*)?$/)) {
      el.value = "auto";
    }
  }

  setSearchLimit();
};

if (navigator.cookieEnabled) {
  window.onresize = setSearchLimit;
}
