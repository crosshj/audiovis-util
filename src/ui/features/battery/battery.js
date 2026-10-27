(function () {
  'use strict';
  var $ = function (id) { return document.getElementById(id); };
  function update(charge, status) {
    var known = charge >= 0, charging = status === 'Charging';
    $('battery-fill').setAttribute('width', known ? 19 * charge / 100 : 0);
    $('battery-bolt').style.display = charging ? 'block' : 'none';
    $('battery-unknown').style.display = known ? 'none' : 'block';
    $('battery').classList.toggle('low', known && charge <= 15 && !charging);
    $('battery').title = known ? charge + '% · ' + status : 'Battery status unavailable';
    $('battery').dataset.charge = charge;
    $('battery-percent').textContent = known ? charge + '%' : '—';
    $('battery-description').textContent = !known ? 'Battery status unavailable' : status === 'Full' ? 'Fully charged' : charging ? 'Charging' : status === 'Discharging' ? 'Running on battery' : status === 'Not charging' ? 'Not charging' : 'Charge status unavailable';
    $('battery-source').textContent = status === 'Discharging' ? 'Battery' : (charging || status === 'Full' || status === 'Not charging') ? 'Power adapter' : 'Unknown';
    $('battery-gauge-fill').style.width = (known ? charge : 0) + '%';
    $('battery-gauge-fill').classList.toggle('low', known && charge <= 15 && !charging);
  }
  popovers.attach($('battery'), $('battery-popover'));
  window.batteryTray = { update: update };
})();
