function sendToast(message) {
  const toast = document.createElement("div");

  toast.classList.add("toast");

  toast.textContent = message;

  document.body.appendChild(toast);

  setTimeout(() => {
    toast.classList.add("visible");
  }, 0);

  setTimeout(() => {
    toast.classList.remove("visible");
    setTimeout(() => {
      toast.remove();
    }, 1500);
  }, 2000);
}

function copyToClipboard(text, message) {
  navigator.clipboard.writeText(text);
  sendToast(message ?? "Copied to clipboard");
}

function q(selector) {
  return document.querySelector(selector);
}

function qAll(selector) {
  return document.querySelectorAll(selector);
}

function c(classes) {
  if (Array.isArray(classes)) {
    return classes.flatMap((i) => c(i)).join(" ");
  }

  if (classes == null || classes == false || classes == true) {
    return "";
  }

  if (typeof classes === "string") {
    return classes;
  }

  if (typeof classes === "object") {
    return Object.entries(classes)
      .filter(([_, value]) => value)
      .map(([key]) => key)
      .join(" ");
  }

  return "";
}

// creates an element. supports:
// parent, element, id, classes
// element, id, classes
// element, classes
// parent, element
// element

function e() {
  let parent = null;
  let element = "div";

  arguments = [...arguments];

  if (arguments[0] instanceof HTMLElement) {
    parent = arguments[0];
    element = arguments[1];
    arguments = arguments.slice(2);
  } else {
    element = arguments[0];
    arguments = arguments.slice(1);
  }

  const e = document.createElement(element);

  if (arguments[1] != null) {
    e.id = arguments[0];
    e.className = arguments[1];
  } else {
    e.className = arguments[0];
  }

  if (parent != null) {
    parent.appendChild(e);
  }

  return e;
}

function humanizeBytes(bytes) {
  if (bytes < 1024) return `${bytes} B`;

  const units = ["KB", "MB", "GB", "TB", "PB"];
  let size = bytes;

  for (const unit of units) {
    size /= 1024;
    if (size < 1024) {
      return `${size.toFixed(1)} ${unit}`;
    }
  }

  return `${size.toFixed(1)} ${units[units.length - 1]}`;
}
