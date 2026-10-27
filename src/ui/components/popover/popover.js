// Shared anchored popovers. Panels live at body level to escape card stacking.
(function () {
  'use strict';
  var active = null;
  function close() {
    if (!active) return false;
    active.panel.hidden = true;
    active.anchor.classList.remove('popover-active');
    active = null;
    return true;
  }
  function position() {
    if (!active) return;
    var anchor = active.anchor.getBoundingClientRect();
    var panel = active.panel;
    var width = document.body.clientWidth;
    var height = document.body.clientHeight;
    var margin = 12, gap = 10;
    // Rects are logical CSS coordinates even when Ultralight uses body zoom.
    var left = Math.max(margin, Math.min(anchor.right - panel.offsetWidth + 8, width - panel.offsetWidth - margin));
    var top = anchor.bottom + gap;
    if (top + panel.offsetHeight > height - margin) top = anchor.top - panel.offsetHeight - gap;
    panel.style.left = left + 'px';
    panel.style.top = Math.max(margin, top) + 'px';
  }
  function attach(anchor, panel) {
    panel.classList.add('popover');
    panel.hidden = true;
    document.body.appendChild(panel);
    var handle = {
      open: function () {
        close();
        active = {anchor: anchor, panel: panel};
        panel.hidden = false;
        anchor.classList.add('popover-active');
        position();
      },
      close: function () { if (active && active.panel === panel) return close(); return false; }
    };
    anchor.addEventListener('click', function (event) {
      event.stopPropagation();
      if (active && active.panel === panel) close(); else handle.open();
    });
    return handle;
  }
  document.addEventListener('mousedown', function (event) {
    if (active && !active.panel.contains(event.target) && !active.anchor.contains(event.target)) close();
  }, true);
  window.addEventListener('resize', position);
  window.popovers = {attach: attach, close: close, reposition: position};
})();
