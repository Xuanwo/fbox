import m from '../../vendor/mithril.js';

export default function Header () {
  return {
    view: () => {
      return m('fb-header',
        m('img', { src: './img/logo.svg' })
      );
    }
  };
}
