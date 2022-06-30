function strike(el) {
  if (el.style.textDecoration) {
    el.style.removeProperty("text-decoration");
  } else {
    el.style.setProperty("text-decoration", "line-through");
  }
}

window.strike = strike;

if ("scrollRestoration" in history) {
  history.scrollRestoration = "auto";
}

function setImageSrc(el, imgID, fallback) {
  const img = document.getElementById(imgID);
  const [file] = el.files;
  if (file) {
    img.src = URL.createObjectURL(file);
  } else {
    img.src = fallback;
  }
}

window.setImageSrc = setImageSrc;
