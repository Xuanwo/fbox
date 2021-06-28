import m from '../vendor/mithril.js'
import Header from './components/Header.js'
import FileList from './components/FileList.js'

document.addEventListener('DOMContentLoaded', function () {
  function render () {
    const ui = m('div',
      m(Header),
      m(FileList)
    );

    m.render(document.body, ui);
  }

  m.redraw = render;
  render();
});
