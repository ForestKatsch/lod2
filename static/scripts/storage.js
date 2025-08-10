function initDragAndDrop() {
  const uploadZone = q("#upload-drop-zone");
  const trashZone = q("#trash-drop-zone");
  let draggingInternalFile = null; // File from within the current directory
  let dragCounter = 0; // Track enter/leave events
  let currentDropTarget = null; // Track which directory we're hovering over
  let isDragging = false; // Track if any drag is active
  let dragTimeout = null; // Timeout to detect abandoned drags

  function warnWhenLeaving() {
    return "File upload will be cancelled.";
  }

  function updateFileTable(html) {
    htmx.swap("#file-table", html, { swapStyle: "outerHTML" });
  }

  function uploadFile(file, targetPath = "{{ .Path }}") {
    q(".empty-directory").remove();

    const eRow = e(q("#file-table > tbody"), "tr", "upload-file");
    const eName = e(eRow, "td", "file-name");
    const eSize = e(eRow, "td", "file-size");
    const eLastModified = e(eRow, "td", "file-last-modified");

    eName.textContent = file.name;
    eSize.textContent = `starting...`;
    eLastModified.textContent = "uploading...";

    const formData = new FormData();
    formData.append("file", file);

    const xhr = new XMLHttpRequest();
    xhr.open("POST", `/files${targetPath}`, true);

    xhr.upload.addEventListener("progress", (e) => {
      if (e.lengthComputable) {
        const progress = e.loaded / e.total;
        const bytes = e.loaded;
        const total = e.total;
        const percentage = Math.round(progress * 100);
        eSize.textContent = `${humanizeBytes(bytes)} / ${humanizeBytes(total)} (${percentage}%)`;
      }
    });

    xhr.addEventListener("load", () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        eName.innerHTML = "";
        const eLink = e(eName, "a", "link file-link");
        // Properly encode the URL path
        const encodedPath = "{{ $.Path }}"
          .split("/")
          .map((p) => (p ? encodeURIComponent(p) : ""))
          .join("/");
        eLink.href = `/files${encodedPath}/${encodeURIComponent(file.name)}`;
        eLink.textContent = file.name;

        eSize.textContent = humanizeBytes(file.size);
        eLastModified.textContent = "just now";
      } else {
        sendToast(`Upload failed: ${xhr.statusText}`);
        eSize.textContent = `failed (${xhr.statusText})`;
        eLastModified.textContent = "just now";
      }
      window.removeEventListener("beforeunload", warnWhenLeaving);
    });

    xhr.addEventListener("error", () => {
      sendToast(`Upload failed: Network error`);
      eSize.textContent = `failed (network error)`;
      eLastModified.textContent = "just now";
      window.removeEventListener("beforeunload", warnWhenLeaving);
    });

    window.addEventListener("beforeunload", warnWhenLeaving);
    xhr.send(formData);
  }

  async function deleteFile(path) {
    const response = await fetch(path, {
      method: "DELETE",
    });

    if (response.ok) {
      updateFileTable(await response.text());
    } else {
      sendToast(`Deletion failed: ${response.statusText}`);
    }
  }

  async function moveFile(sourcePath, destPath) {
    console.log("Moving file:", sourcePath, "to:", destPath);

    // Extract just the pathname from the full URL and construct the API URL
    const sourceUrl = new URL(sourcePath);
    const apiPath = sourceUrl.pathname; // This gives us just /files/path/to/file.ext

    // Create the PATCH request URL
    const patchUrl = new URL(apiPath, window.location.origin);
    patchUrl.searchParams.set("dest", destPath);

    const response = await fetch(patchUrl.toString(), {
      method: "PATCH",
    });

    console.log("Move response status:", response.status, response.statusText);

    if (response.ok) {
      updateFileTable(await response.text());
      sendToast(`Moved successfully`);
    } else {
      const errorText = await response.text();
      console.log("Move error response:", errorText);
      sendToast(`Move failed: ${response.statusText} - ${errorText}`);
    }
  }

  function hideAllZones() {
    uploadZone.classList.remove("visible");
    trashZone.classList.remove("visible");
    // Remove highlighting from directory drop targets
    document
      .querySelectorAll(".directory-drop-target.drag-over")
      .forEach((el) => {
        el.classList.remove("drag-over");
      });
    currentDropTarget = null;
    isDragging = false;
    dragCounter = 0;
    // Clear any pending timeout
    if (dragTimeout) {
      clearTimeout(dragTimeout);
      dragTimeout = null;
    }
  }

  // Handle dragstart for internal files
  document.addEventListener("dragstart", (e) => {
    const fileLink = e.target.closest(".file-link");
    if (fileLink) {
      draggingInternalFile = fileLink;
      e.dataTransfer.effectAllowed = "move";
    }
  });

  // Global dragenter - detect what kind of drag this is
  document.addEventListener("dragenter", (e) => {
    dragCounter++;
    isDragging = true;

    // Check if this is an external file drag
    const hasFiles =
      e.dataTransfer.types && e.dataTransfer.types.includes("Files");

    if (hasFiles && !draggingInternalFile) {
      // External file drag - show upload zone and highlight directory targets
      uploadZone.classList.add("visible");

      // Check if we're entering a directory drop target for external files
      const dropTarget = e.target.closest(".directory-drop-target");
      if (dropTarget && dropTarget !== currentDropTarget) {
        // Remove highlight from previous target
        if (currentDropTarget) {
          currentDropTarget.classList.remove("drag-over");
        }
        // Add highlight to new target
        dropTarget.classList.add("drag-over");
        currentDropTarget = dropTarget;
      }
    } else if (draggingInternalFile) {
      // Internal file drag - show trash zone and highlight directory targets
      trashZone.classList.add("visible");

      // Check if we're entering a directory drop target
      const dropTarget = e.target.closest(".directory-drop-target");
      if (dropTarget && dropTarget !== currentDropTarget) {
        // Remove highlight from previous target
        if (currentDropTarget) {
          currentDropTarget.classList.remove("drag-over");
        }
        // Add highlight to new target
        dropTarget.classList.add("drag-over");
        currentDropTarget = dropTarget;
      }
    }

    e.preventDefault();
  });

  // Global dragleave - hide zones when leaving document
  document.addEventListener("dragleave", (e) => {
    dragCounter--;

    // Only hide when we've left the document entirely
    if (dragCounter === 0) {
      // Set a timeout to hide zones if drag doesn't return quickly
      if (dragTimeout) clearTimeout(dragTimeout);
      dragTimeout = setTimeout(() => {
        if (dragCounter <= 0) {
          hideAllZones();
        }
      }, 100); // Small delay to handle rapid enter/leave events
    }
  });

  // Global dragover - keep zones visible and prevent default
  document.addEventListener("dragover", (e) => {
    // Clear any pending timeout since we're still dragging
    if (dragTimeout) {
      clearTimeout(dragTimeout);
      dragTimeout = null;
    }
    e.preventDefault();
  });

  // Handle mouse leaving the window during drag
  document.addEventListener("mouseleave", (e) => {
    if (isDragging && !draggingInternalFile) {
      // Only hide for external file drags when mouse leaves window
      hideAllZones();
    }
  });

  // Handle ESC key when zones are visible (for keyboard accessibility)
  document.addEventListener("keydown", (e) => {
    if (
      e.key === "Escape" &&
      (uploadZone.classList.contains("visible") ||
        trashZone.classList.contains("visible"))
    ) {
      hideAllZones();
    }
  });

  // Global dragend - cleanup when drag operation ends
  document.addEventListener("dragend", (e) => {
    hideAllZones();
    draggingInternalFile = null;
    dragCounter = 0;
  });

  // Upload zone drop handler
  uploadZone.addEventListener("drop", (e) => {
    e.preventDefault();
    hideAllZones();
    dragCounter = 0;

    if (e.dataTransfer.files.length > 0) {
      // Upload each file
      Array.from(e.dataTransfer.files).forEach((file) => {
        uploadFile(file);
      });
    }
  });

  // Trash zone drop handler
  trashZone.addEventListener("drop", (e) => {
    e.preventDefault();
    hideAllZones();
    dragCounter = 0;

    if (draggingInternalFile) {
      const href = draggingInternalFile.href;
      deleteFile(href);
      draggingInternalFile = null;
    }
  });

  // Handle drops on directory targets and page
  document.addEventListener("drop", (e) => {
    const dropTarget = e.target.closest(".directory-drop-target");
    const hasFiles = e.dataTransfer.files && e.dataTransfer.files.length > 0;

    // Handle internal file drops on directories
    if (dropTarget && draggingInternalFile) {
      e.preventDefault();
      hideAllZones();
      dragCounter = 0;

      const sourcePath = draggingInternalFile.href;
      const destPath = dropTarget.dataset.path;

      moveFile(sourcePath, destPath);
      draggingInternalFile = null;
      return;
    }

    // Handle external file drops on directories
    if (dropTarget && hasFiles) {
      e.preventDefault();
      hideAllZones();
      dragCounter = 0;

      const targetPath = dropTarget.dataset.path;
      Array.from(e.dataTransfer.files).forEach((file) => {
        uploadFile(file, targetPath);
      });
      return;
    }

    // Handle external file drops anywhere on page (fallback to current directory)
    if (
      hasFiles &&
      !e.target.closest("#upload-drop-zone") &&
      !e.target.closest("#trash-drop-zone")
    ) {
      e.preventDefault();
      hideAllZones();
      dragCounter = 0;

      Array.from(e.dataTransfer.files).forEach((file) => {
        uploadFile(file); // Uses default current directory
      });
      return;
    }

    // Prevent default drop behavior on document for other cases
    if (
      !e.target.closest("#upload-drop-zone") &&
      !e.target.closest("#trash-drop-zone")
    ) {
      e.preventDefault();
    }
  });

  q("#create-directory-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    const directoryName = formData.get("directoryName");
    try {
      const response = await fetch(`/files{{ .Path }}/${directoryName}`, {
        method: "PUT",
      });

      updateFileTable(await response.text());
      q("#create-directory-form").reset();
    } catch (e) {
      sendToast(`Directory creation failed: ${e.message}`);
    }
  });
}

document.addEventListener("DOMContentLoaded", initDragAndDrop);
