let currentGalleryID = null;
let currentGalleryIndex = 0;

function openModal(galleryId, index) {
  currentGalleryID = galleryId;
  currentGalleryIndex = index;
  const images = document
    .querySelector(`.gallery[data-gallery-id="${galleryId}"]`)
    .querySelectorAll("img");
  document.getElementById("modal").style.display = "block";
  document.getElementById("modalImage").src = images[index].src;
  document.addEventListener("keydown", handleKeyPress);
  document.querySelector(".close").addEventListener("click", closeModal);
  document
    .querySelector(".modal-content")
    .addEventListener("click", function (e) {
      e.stopPropagation();
    });
}

function changeImage(change) {
  const images = document
    .querySelector(`.gallery[data-gallery-id="${currentGalleryID}"]`)
    .querySelectorAll("img");

  currentGalleryIndex += change;
  if (currentGalleryIndex < 0) currentGalleryIndex = images.length - 1;
  else if (currentGalleryIndex >= images.length) currentGalleryIndex = 0;

  document.getElementById("modalImage").src = images[currentGalleryIndex].src;
}

function closeModal(event) {
  const modalContent = document.querySelector(".modal-content");
  if (
    event.target === document.getElementById("modal") ||
    !modalContent.contains(event.target)
  ) {
    document.getElementById("modal").style.display = "none";
    document.removeEventListener("keydown", handleKeyPress);
    document.getElementById("modal").removeEventListener("click", closeModal);
    currentGalleryID = null;
  }
}

function handleKeyPress(event) {
  if (event.key === "Escape") {
    closeModal(event);
  }
  if (event.key === "ArrowLeft") {
    changeImage(-1);
  }
  if (event.key === "ArrowRight") {
    changeImage(1);
  }
}
