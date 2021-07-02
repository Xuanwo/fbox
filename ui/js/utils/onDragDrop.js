export default function onDragDrop (handleDrop) {
  const dropZone = document.body;

  dropZone.addEventListener('dragover', function (event) {
    event.stopPropagation();
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
  });

  dropZone.addEventListener('drop', function (event) {
    event.stopPropagation();
    event.preventDefault();
    const files = event.dataTransfer.files;
    handleDrop(files);
  });
}
