// Reusable modal host. Content stays owned by the caller; sizing/actions are shared.
(function () {
  'use strict';
  var active = null;
  var host = document.createElement('div');
  host.className = 'modal-backdrop'; host.hidden = true;
  document.body.appendChild(host);
  function close(force) {
    if (!active) return false;
    if (active.busy && !force) return true;
    var previous = active;
    previous.content.hidden = true;
    document.body.appendChild(previous.content);
    host.innerHTML = ''; host.hidden = true; active = null;
    if (previous.options.onClose) previous.options.onClose();
    return true;
  }
  function open(options) {
    if (active && active.busy) return null;
    close(true); popovers.close();
    var panel = document.createElement('div'); panel.className = 'modal-panel';
    var widths = {small:360, medium:480, large:680, wide:920};
    panel.style.width = (typeof options.size === 'number' ? Math.max(280,options.size) : widths[options.size] || widths.medium) + 'px';
    var header = document.createElement('div'); header.className = 'modal-header';
    var title = document.createElement('div'); title.className = 'modal-title'; title.textContent = options.title;
    var dismiss = document.createElement('div'); dismiss.className = 'modal-close'; dismiss.textContent = '×'; dismiss.onclick = function(){close(false)};
    header.appendChild(title); header.appendChild(dismiss); panel.appendChild(header);
    if (options.description) {var description=document.createElement('div');description.className='modal-description';description.textContent=options.description;panel.appendChild(description);}
    var body = document.createElement('div'); body.className = 'modal-body';
    options.content.hidden=false; body.appendChild(options.content);panel.appendChild(body);
    var footer=document.createElement('div');footer.className='modal-footer';
    active={options:options,content:options.content,busy:false};
    var handle={close:function(){return close(false)},setBusy:function(value){if(active&&active.content===options.content){active.busy=value;panel.classList.toggle('busy',value);}}};
    (options.actions||[]).forEach(function(action){var control=document.createElement('div');control.className='modal-action'+(action.primary?' primary':'');control.textContent=action.label;if(action.id)control.id=action.id;control.onclick=function(){if(active&&!active.busy)action.run(handle)};footer.appendChild(control);});
    panel.appendChild(footer);host.appendChild(panel);host.hidden=false;
    return handle;
  }
  host.addEventListener('mousedown',function(event){if(event.target===host&&active&&active.options.dismissOnBackdrop)close(false)});
  window.modals={open:open,close:function(){return close(false)},escape:function(){return close(false)},isOpen:function(){return !!active}};
})();
