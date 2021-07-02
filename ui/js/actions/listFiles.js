export default async function listFiles (context) {
  const response = await window.fetch('/files');
  const json = await response.json();

  const files = Object
    .keys(json)
    .reduce((files, file) => {
      files.push(json[file]);
      return files;
    }, []);

  context.files = files;
  context.redraw();

  return files;
}
