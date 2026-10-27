(function () {
  'use strict';
  var $ = function (id) { return document.getElementById(id); };
  var wifiState = { available: false, enabled: false }, wifiPending = false;
  function update(data) {
    wifiState = data; wifiPending = false;
    $('wifi-toggle').classList.toggle('on', data.enabled);
    $('wifi-toggle').classList.toggle('disabled', !data.available || data.blocked);
    $('wifi-toggle').classList.remove('pending');
    $('wifi').classList.toggle('wifi-off', !data.enabled);
    $('wifi-status').textContent = !data.available ? 'Wi-Fi unavailable' : data.blocked ? 'Wi-Fi blocked by hardware' : !data.enabled ? 'Off' : data.connected ? 'On' : 'Connecting…';
    $('wifi-connection').hidden = !data.enabled;
    $('wifi-network-name').textContent = data.connected ? data.ssid : 'Not connected';
    $('wifi-network-detail').textContent = data.connected ? 'Connected' : 'Waiting for the saved network';
    $('wifi-check').hidden = !data.connected;
    $('wifi-error').hidden = !data.error; $('wifi-error').textContent = data.error || '';
    popovers.reposition();
  }
  $('wifi-toggle').onclick = function (event) {
    event.stopPropagation();
    if (wifiPending || !wifiState.available || wifiState.blocked) return;
    var enabled = !wifiState.enabled;
    if (App.command('wifi.setEnabled', enabled ? 1 : 0) === -1) {
      $('wifi-error').hidden = false;
      $('wifi-error').textContent = 'Could not change Wi-Fi. Try again.';
      return;
    }
    wifiPending = true;
    $('wifi-toggle').classList.add('pending');
    $('wifi-status').textContent = enabled ? 'Turning on…' : 'Turning off…';
  };
  popovers.attach($('wifi'), $('wifi-popover'));
  window.wifiTray = { update: update };
})();
