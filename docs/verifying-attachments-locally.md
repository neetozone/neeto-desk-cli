# Verifying attachments against a local server

`--attach` and `attachments upload` cannot be exercised by the unit tests alone: they
reserve an upload through the API and then send the bytes to a second host. This is how to
run the whole path against a local `neeto-desk-web`.

## 1. Run the web app on port 8910

```bash
cd ../neeto-desk-web
bin/rails s -p 8910
yarn dev
```

Port 8910 is not arbitrary. The reservation builds `upload_url` from the workspace's own
root URL, which `.env.development` pins to `<subdomain>.lvh.me:8910`. On any other port the
reservation succeeds and the `PUT` then goes to a port nothing is listening on.

`lvh.me` resolves to 127.0.0.1 for every subdomain, so no `/etc/hosts` entry is needed.

## 2. Mint a CLI session

The normal `neetodesk login` needs a browser round trip. For a local check, create the
session directly and print its token:

```bash
cd ../neeto-desk-web
bin/rails runner '
  org = Organization.find_by(subdomain: "spinkart")
  user = org.users.first
  token = SecureRandom.hex(24)
  NeetoCommonsBackend::CliSession.create!(
    organization: org, user:, email: user.email, status: :authenticated,
    session_token: token, authenticated_at: Time.current)
  puts token
'
```

## 3. Point the CLI at it, without touching your real credentials

Write the session into a throwaway `HOME` so `~/.config/neetodesk/auth.json` is left alone:

```bash
export HOME=/tmp/neetodesk-verify
mkdir -p "$HOME/.config/neetodesk"
cat > "$HOME/.config/neetodesk/auth.json" <<'JSON'
{"credentials":[{"subdomain":"spinkart","email":"<user email>","session_token":"<token>"}]}
JSON

export NEETODESK_BASE_URL=http://spinkart.lvh.me:8910
go build -o /tmp/neetodesk ./cmd/neetodesk/
```

## 4. Upload a real image

Use a genuine image, not a text file with an image extension — the server identifies the
file and records its dimensions, so a fake one passes the upload and then fails to analyse.

```bash
magick -size 800x600 gradient:navy-orange /tmp/printer-error.png

/tmp/neetodesk attachments upload /tmp/printer-error.png

/tmp/neetodesk tickets create \
  --email customer@example.com \
  --subject "Printer is jammed" \
  --description "Photo attached." \
  --attach /tmp/printer-error.png

/tmp/neetodesk tickets comments create <ticket-number> \
  --content "<p>Close-up of the jam.</p>" \
  --attach /tmp/printer-error.png
```

## 5. Confirm what landed

Check the content type, the recorded dimensions and the bytes:

```bash
cd ../neeto-desk-web
bin/rails runner '
  comment = Organization.find_by(subdomain: "spinkart").tickets.find_by(number: <n>).description_comment
  comment.attachments.each do |a|
    b = a.blob
    b.analyze unless b.analyzed?
    puts [b.filename, b.content_type, b.byte_size, b.metadata.slice("width", "height")].inspect
  end
'
```

`identified: true` in the metadata means the server actually parsed the image. To prove the
bytes survived, download the attachment and compare it with the file you sent:

```bash
shasum -a 256 /tmp/printer-error.png
curl -sL -o /tmp/served.png "<the blob URL>" && shasum -a 256 /tmp/served.png
```

## What the size limits should do

Two different ceilings apply, and a file between them is meant to fail at the second one:

| File size | Where it stops | Message |
| --- | --- | --- |
| Up to 10 MB | Attaches | — |
| 10-50 MB | Reserved and sent, refused at attach | `Attachments must be 10 MB or smaller.` |
| Over 50 MB | Refused at reservation, no bytes sent | `File must be 50 MB or smaller.` |

Both cases exit non-zero and create no ticket. A file that was uploaded before a later one
failed is left unattached and reclaimed by the server's cleanup, so re-running is safe.
