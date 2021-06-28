export default async function listFiles () {
  const response = await fetch('/api/files');
  const json = await response.json();

  const files = Object
    .keys(json)
    .reduce((files, file) => {
      files.push(json[file]);
      return files;
    }, [])

  return files;
}
