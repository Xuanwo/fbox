import m from '../../vendor/mithril.js'

export default function Header () {
  function handleChange (onChange) {
    return (event) => {
      onChange(event.target.files);
    }
  }

  return {
    view: (vnode) => {
      return m('fb-fileuploader',
        m('label', { for: 'file-uploader' }, vnode.children),
        m('input', {
          id: 'file-uploader',
          onchange: handleChange(vnode.attrs.onChange),
          type: 'file'
        })
      )
    }
  }
}
