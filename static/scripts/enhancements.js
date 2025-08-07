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
