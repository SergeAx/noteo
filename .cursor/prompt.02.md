Now, let's handle our subscriptions management. We need to offer user those
actions for each subscription:

1) silence notifications
2) mute notifications
3) unsubscribe from notifications

In case of silence and mute we should provide 4 options: for 30 minutes, for
2 hours, until the end of the day and for an arbitrary time specified by the
user.

In case of unsubscribe we should provide "Subscribe again" button after
confirmation.

Handling of silence and mute buttons should be done by a separate handler, same
for both cases.

If notifications are already silenced or muted, "Unmute" and "Unsilence"
buttons should be available instead of "Mute" and "Silence".
