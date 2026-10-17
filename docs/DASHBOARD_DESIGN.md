# HaloMail account interface

The signed-in workspace uses a text-led masthead, open spacing, and a narrow
connection column. The account label is simply “Personal workspace.”
Hierarchy comes from alignment and spacing, with quiet fills for table headings
and connection tools. Dividers are reserved for the masthead and form controls. Geist and Geist Mono remain the only fonts. The palette uses
warm paper and ink in light mode, charcoal and off-white in dark mode, with the
existing violet reserved for small status accents.

There are no decorative icons, gradients, glass panels, floating cards, or
promotional slogans in the account UI. All styles live in the dashboard route
group. Marketing, login, registration, public booking, and checkout logic are
unchanged.

## Product references

Screenshots and workflow documentation were reviewed directly:

- [Cal.com bookings](https://cal.com/blog/a-complete-walkthrough-of-cal-com-s-booking-dashboard-its-key-features):
  date-led booking records and compact actions.
- [Fillout responses](https://www.fillout.com/help/view-responses):
  aligned columns, readable response records, and expandable details.
- [Tally workspaces](https://tally.so/help/workspaces):
  clear organization and straightforward navigation.

HaloMail's layout is specific to its two workflows. Forms puts submissions first,
with access keys, allowances, and embed code alongside. Meetings puts bookings
first, with calendar connection, event types, and availability in an expandable
Scheduling setup section. Direct links to setup open it automatically. Settings
uses labeled sections, and Billing uses plan comparison rows.

## Behavior

The existing RPCs, payloads, key lifecycle, calendar flow, event creation,
availability rules and overrides, profile saves, and payment actions are retained.
Overview displays existing account data without fabricated trends or metrics.
Loading, empty, and failed requests have distinct presentations.

Shared account state keeps profile edits in sync and refreshes allowances following
billing events. List previews show only the sender, message, and date; opening a
submission reveals the sender email, timestamp, and full form fields. Help text
appears beside relevant actions or inside the embed disclosure. One-time key values stay copyable when the
follow-up list request fails and are masked after leaving the page.

## Verification

The isolated browser review intercepts all account APIs with test fixtures. It
checks desktop and mobile overflow, theme switching, navigation, submission
disclosures, stable usage refresh, event and availability payloads, profile saves,
one-time key masking, and failed requests. Desktop light/dark and mobile screen
captures were inspected. This is not a live payment or Google OAuth test.
