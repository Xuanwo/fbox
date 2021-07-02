export default async function uploadFile (file) {
  const response = await window.fetch('/upload/' + file.name, {
    method: 'post',
    body: file
  });

  return response;
}
