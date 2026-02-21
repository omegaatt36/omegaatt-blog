/**
 * Gallery Component — PhotoSwipe v5 wrapper
 */
import PhotoSwipe from "./photoswipe.esm.min.js";
import PhotoSwipeLightbox from "./photoswipe-lightbox.esm.min.js";

document.querySelectorAll("[data-pswp-gallery]").forEach((gallery) => {
  const lightbox = new PhotoSwipeLightbox({
    gallery,
    children: "a[data-pswp-width]",
    pswpModule: PhotoSwipe,
    preload: [1, 2],
    pinchToClose: true,
    closeOnVerticalDrag: true,
  });
  lightbox.init();
});
