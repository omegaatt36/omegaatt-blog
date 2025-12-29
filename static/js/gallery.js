/**
 * Gallery Component
 * A lightweight, accessible image gallery with modal viewer
 */
(function () {
  "use strict";

  // State
  let currentGalleryId = null;
  let currentIndex = 0;
  let galleryImages = [];
  let touchStartX = 0;
  let touchEndX = 0;
  let isAnimating = false;

  // DOM Elements (cached after init)
  let modal = null;
  let modalImage = null;
  let modalCaption = null;
  let modalSpinner = null;
  let modalCounter = null;
  let modalCurrent = null;
  let modalTotal = null;
  let prevBtn = null;
  let nextBtn = null;
  let closeBtn = null;

  // Minimum swipe distance for navigation (in pixels)
  const SWIPE_THRESHOLD = 50;

  /**
   * Initialize the gallery component
   */
  function init() {
    // Cache DOM elements
    modal = document.getElementById("gallery-modal");
    if (!modal) {
      // No gallery modal on this page, skip initialization
      return;
    }

    modalImage = modal.querySelector(".gallery-modal-image");
    modalCaption = modal.querySelector(".gallery-modal-caption");
    modalSpinner = modal.querySelector(".gallery-modal-spinner");
    modalCounter = modal.querySelector(".gallery-modal-counter");
    modalCurrent = modal.querySelector(".gallery-modal-current");
    modalTotal = modal.querySelector(".gallery-modal-total");
    prevBtn = modal.querySelector(".gallery-modal-prev");
    nextBtn = modal.querySelector(".gallery-modal-next");
    closeBtn = modal.querySelector(".gallery-modal-close");

    // Bind events using event delegation for gallery images
    document.addEventListener("click", handleDocumentClick);

    // Modal event listeners
    closeBtn.addEventListener("click", closeModal);
    prevBtn.addEventListener("click", () => navigate(-1));
    nextBtn.addEventListener("click", () => navigate(1));
    modal.addEventListener("click", handleModalClick);

    // Keyboard navigation
    document.addEventListener("keydown", handleKeydown);

    // Touch events for swipe navigation
    modal.addEventListener("touchstart", handleTouchStart, { passive: true });
    modal.addEventListener("touchend", handleTouchEnd, { passive: true });

    // Image load event
    modalImage.addEventListener("load", handleImageLoad);
    modalImage.addEventListener("error", handleImageError);

    // Preload adjacent images when modal is open
    modalImage.addEventListener("load", preloadAdjacentImages);
  }

  /**
   * Handle clicks on the document (event delegation)
   */
  function handleDocumentClick(event) {
    const galleryImage = event.target.closest(".gallery-image");
    if (!galleryImage) return;

    const gallery = galleryImage.closest(".gallery");
    if (!gallery) return;

    event.preventDefault();

    const galleryId = gallery.dataset.galleryId;
    const images = Array.from(gallery.querySelectorAll(".gallery-image"));
    const index = images.indexOf(galleryImage);

    openModal(galleryId, images, index);
  }

  /**
   * Open the modal with specified gallery and image index
   */
  function openModal(galleryId, images, index) {
    if (!modal || isAnimating) return;

    currentGalleryId = galleryId;
    galleryImages = images.map((img) => ({
      src: img.dataset.fullSrc || img.src,
      alt: img.alt || "",
      caption: img.dataset.caption || "",
    }));
    currentIndex = index;

    // Update counter
    updateCounter();

    // Show modal
    modal.classList.add("active");
    modal.setAttribute("aria-hidden", "false");
    document.body.classList.add("gallery-modal-open");

    // Load the image
    loadImage(currentIndex);

    // Set focus to modal for accessibility
    closeBtn.focus();

    // Store the element that had focus before opening
    modal.dataset.previousFocus =
      document.activeElement?.id || document.activeElement?.className || "";
  }

  /**
   * Close the modal
   */
  function closeModal() {
    if (!modal || isAnimating) return;

    isAnimating = true;
    modal.classList.add("closing");

    setTimeout(() => {
      modal.classList.remove("active", "closing");
      modal.setAttribute("aria-hidden", "true");
      document.body.classList.remove("gallery-modal-open");

      // Reset state
      modalImage.src = "";
      modalImage.alt = "";
      modalImage.classList.remove("loaded");
      modalCaption.textContent = "";
      currentGalleryId = null;
      galleryImages = [];
      currentIndex = 0;
      isAnimating = false;

      // Return focus to the gallery
      const gallery = document.querySelector(
        `.gallery[data-gallery-id="${currentGalleryId}"]`,
      );
      if (gallery) {
        const images = gallery.querySelectorAll(".gallery-image");
        if (images[currentIndex]) {
          images[currentIndex].focus();
        }
      }
    }, 300);
  }

  /**
   * Navigate to previous or next image
   */
  function navigate(direction) {
    if (isAnimating || galleryImages.length <= 1) return;

    let newIndex = currentIndex + direction;

    // Wrap around
    if (newIndex < 0) {
      newIndex = galleryImages.length - 1;
    } else if (newIndex >= galleryImages.length) {
      newIndex = 0;
    }

    currentIndex = newIndex;
    updateCounter();
    loadImage(currentIndex);
  }

  /**
   * Load an image by index
   */
  function loadImage(index) {
    const imageData = galleryImages[index];
    if (!imageData) return;

    // Show spinner, hide image
    modalSpinner.classList.remove("hidden");
    modalImage.classList.remove("loaded");

    // Update image
    modalImage.src = imageData.src;
    modalImage.alt = imageData.alt;

    // Update caption
    modalCaption.textContent = imageData.caption;

    // Update navigation button states
    updateNavigationState();
  }

  /**
   * Handle successful image load
   */
  function handleImageLoad() {
    modalSpinner.classList.add("hidden");
    modalImage.classList.add("loaded");
  }

  /**
   * Handle image load error
   */
  function handleImageError() {
    modalSpinner.classList.add("hidden");
    modalImage.alt = "Failed to load image";
    console.error("Gallery: Failed to load image:", modalImage.src);
  }

  /**
   * Update the image counter display
   */
  function updateCounter() {
    if (modalCurrent && modalTotal) {
      modalCurrent.textContent = currentIndex + 1;
      modalTotal.textContent = galleryImages.length;
    }

    // Hide counter for single images
    if (modalCounter) {
      modalCounter.style.display = galleryImages.length > 1 ? "" : "none";
    }
  }

  /**
   * Update navigation button states
   */
  function updateNavigationState() {
    const singleImage = galleryImages.length <= 1;
    prevBtn.style.display = singleImage ? "none" : "";
    nextBtn.style.display = singleImage ? "none" : "";
  }

  /**
   * Preload adjacent images for smoother navigation
   */
  function preloadAdjacentImages() {
    if (galleryImages.length <= 1) return;

    const prevIndex =
      currentIndex === 0 ? galleryImages.length - 1 : currentIndex - 1;
    const nextIndex =
      currentIndex === galleryImages.length - 1 ? 0 : currentIndex + 1;

    // Preload previous
    if (galleryImages[prevIndex]) {
      const prevImg = new Image();
      prevImg.src = galleryImages[prevIndex].src;
    }

    // Preload next
    if (galleryImages[nextIndex]) {
      const nextImg = new Image();
      nextImg.src = galleryImages[nextIndex].src;
    }
  }

  /**
   * Handle clicks on the modal (close when clicking backdrop)
   */
  function handleModalClick(event) {
    // Close only if clicking the backdrop, not the content
    if (
      event.target === modal ||
      event.target.classList.contains("gallery-modal-content")
    ) {
      closeModal();
    }
  }

  /**
   * Handle keyboard navigation
   */
  function handleKeydown(event) {
    if (!modal || !modal.classList.contains("active")) return;

    switch (event.key) {
      case "Escape":
        event.preventDefault();
        closeModal();
        break;
      case "ArrowLeft":
        event.preventDefault();
        navigate(-1);
        break;
      case "ArrowRight":
        event.preventDefault();
        navigate(1);
        break;
      case "Tab":
        // Trap focus within modal
        handleTabKey(event);
        break;
    }
  }

  /**
   * Trap focus within the modal for accessibility
   */
  function handleTabKey(event) {
    const focusableElements = modal.querySelectorAll(
      'button:not([disabled]), [tabindex]:not([tabindex="-1"])',
    );
    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];

    if (event.shiftKey) {
      // Shift + Tab
      if (document.activeElement === firstElement) {
        event.preventDefault();
        lastElement.focus();
      }
    } else {
      // Tab
      if (document.activeElement === lastElement) {
        event.preventDefault();
        firstElement.focus();
      }
    }
  }

  /**
   * Handle touch start for swipe detection
   */
  function handleTouchStart(event) {
    touchStartX = event.changedTouches[0].screenX;
  }

  /**
   * Handle touch end for swipe detection
   */
  function handleTouchEnd(event) {
    touchEndX = event.changedTouches[0].screenX;
    handleSwipe();
  }

  /**
   * Process swipe gesture
   */
  function handleSwipe() {
    const swipeDistance = touchEndX - touchStartX;

    if (Math.abs(swipeDistance) < SWIPE_THRESHOLD) return;

    if (swipeDistance > 0) {
      // Swipe right - show previous
      navigate(-1);
    } else {
      // Swipe left - show next
      navigate(1);
    }
  }

  // Initialize when DOM is ready
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }

  // Expose minimal API for external use if needed
  window.GalleryModal = {
    open: function (galleryId, index) {
      const gallery = document.querySelector(
        `.gallery[data-gallery-id="${galleryId}"]`,
      );
      if (gallery) {
        const images = Array.from(gallery.querySelectorAll(".gallery-image"));
        openModal(galleryId, images, index || 0);
      }
    },
    close: closeModal,
  };
})();
