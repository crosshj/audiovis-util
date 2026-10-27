// Single JS ↔ Go seam. Features subscribe by topic; Go never builds feature-specific scripts.
(function () {
  'use strict';
  function pad(value) { return ('0' + value).slice(-2); }
  function formatClock(clock) {
    return String(clock.year).slice(2) + pad(clock.month) + pad(clock.day) + '  ' + pad(clock.hour) + ':' + pad(clock.minute);
  }
  window.App = {
    command: function (name, value) {
      return typeof nativeInvoke === 'function' ? nativeInvoke(name, value) : value;
    },
    receive: function (topic, payload) {
      switch (topic) {
        case 'clock.tick':
          if (window.dateTime) dateTime.update(payload);
          if (!document.body.dataset.static && window.dashboard) dashboard.tick(formatClock(payload));
          break;
        case 'clock.saved':
          if (window.dateTime) dateTime.saved(payload);
          break;
        case 'wifi.status':
          if (window.wifiTray) wifiTray.update(payload);
          break;
        case 'battery.status':
          if (window.batteryTray) batteryTray.update(payload.charge, payload.status);
          break;
      }
    }
  };
})();
