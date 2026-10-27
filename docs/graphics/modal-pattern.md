# Shared modal pattern

The Ultralight experiment uses `modal.js` for centered dialogs. Date/time is the
first consumer; the host is independent of its fields and save operation.

```js
var dialog = modals.open({
  title: 'Adjust Date / Time',
  description: 'Time zone: UTC',
  size: 'medium',
  content: document.getElementById('date-editor'),
  dismissOnBackdrop: false,
  onClose: function () { /* discard local draft state */ },
  actions: [
    {label: 'Cancel', run: function (dialog) { dialog.close(); }},
    {label: 'Save', primary: true, run: function (dialog) {
      dialog.setBusy(true);
      // Submit async work. On completion, clear busy, then close or show errors.
    }}
  ]
});
```

Sizing presets are `small` (360px), `medium` (480px), `large` (680px), and
`wide` (920px), in logical CSS pixels. A numeric `size` sets a custom width,
with a minimum of 280px. Width and height are constrained by the viewport;
the body scrolls while the header/footer stay visible. Height follows content.
The overlay centers with flex layout rather than transformed positioning.

The shared shell provides the dark surface, outline, rounded corners, shadow,
dimmed background, close control, action bar and primary/secondary buttons.
Only one modal is active. Opening closes tray popovers. Escape, Cancel and the
close control dismiss; backdrop dismissal is opt-in. Busy state blocks repeat
submissions, editing and dismissal until the operation finishes. This prevents
losing an in-flight save. Underlying pointer interactions are blocked by the
full-screen overlay. The date editor keeps errors inline and stays open if saving
fails. No native form controls or accessibility-specific behavior were added.

Content remains caller-owned: the host mounts it into its scrollable body and
returns it, hidden, to document.body on close. The caller initializes its draft
on every opening. The modal host never knows how an individual form saves data.

## Calendar and clock editor

Clicking the tray date/time opens the current month with today highlighted.
There are no month navigation controls. The calendar is driven by Go's current
local date, so a successful clock change updates both it and the tray clock.
`Adjust Date / Time` opens the medium modal. Custom editable divs and stepper
arrows accept year/month/day and 24-hour hour/minute/second values. The supported
year range is 2000–2099; both JavaScript and Go validate dates, including leap days.
The current timezone is shown (UTC on this appliance); timezone editing is outside
this basic editor.

Save uses a bounded Go worker queue and `timedatectl --no-ask-password set-time`
with an eight-second timeout. It runs off the renderer thread, reports failures
back to the open modal, and closes only after success. Commands use structured
arguments, not shell interpolation. The documented systemd operation updates
both the system clock and hardware RTC:
[systemd timedatectl set-time](https://www.freedesktop.org/software/systemd/man/latest/timedatectl.html#%20set-time%20%5BTIME%5D).

The target has `/dev/rtc0`, a UTC battery-backed clock, and no NTP service.
Persistence uses that advancing hardware clock, not a timestamp file that would
reset the clock on each boot. This works with the RAM-only root filesystem.
A future automatic time-sync service could override manually selected time;
this implementation does not silently change that policy.

## Validation

Actual Ultralight mouse and keyboard tests passed at 1280×800 and 2560×1600 for
calendar opening, today's marker, centered modal bounds, text editing,
impossible-date rejection and Cancel. Backend checks reject invalid leap days,
months and hours and accept February 29, 2024. The normal self-test never changes
the system clock.

A separate hardware Save test submitted the current date/time. The operation
succeeded; afterward system and RTC readings differed by one second. No reboot
was performed as part of this test.

- [Save and RTC verification](validation/datetime-save-test.log)
- [UI validation](validation/datetime-ui-test.log)
- [Calendar](validation/calendar-popover.png)
- [Modal](validation/datetime-modal.png)
