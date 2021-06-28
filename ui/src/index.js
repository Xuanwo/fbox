const m = require('../vendor/mithril.min.js')

document.addEventListener('DOMContentLoaded', function () {
  const ui = m('div', 'hello world');

  m.render(document.body, ui);
});
