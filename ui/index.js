import m from '../vendor/mithril.js'

document.addEventListener('DOMContentLoaded', function () {
  const ui = m('div', 'hello world');

  m.render(document.body, ui);
});
