#!/usr/bin/env python3
"""
Minimal fake IMAP server for development/testing.
Accepts any login, serves fake emails from a local mailbox.

Usage:
    python3 scripts/fake_imap_server.py [--port 1143] [--inject N]

    --port PORT    IMAP port (default: 1143)
    --inject N     Generate N fake emails on startup

Inject more emails while running:
    echo "inject 5" | nc localhost 1144

Emails can also be placed as .eml files in ./fake_mailbox/
"""

import argparse
import os
import random
import socket
import threading
import time
import email.utils
import uuid
import signal
import sys

MAILBOX_DIR = os.environ.get("MAILBOX_DIR", os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "fake_mailbox"))
CONTROL_PORT_OFFSET = 1  # control port = imap_port + 1

uid_counter = 0
uid_lock = threading.Lock()

def next_uid():
    global uid_counter
    with uid_lock:
        uid_counter += 1
        return uid_counter

def ensure_mailbox():
    os.makedirs(MAILBOX_DIR, exist_ok=True)

def inject_fake_emails(count):
    ensure_mailbox()
    senders = [
        ("Jan Kowalski", "jan.kowalski@example.com"),
        ("Anna Nowak", "anna.nowak@firma.pl"),
        ("Support Team", "support@megacorp.com"),
        ("Maria Wiśniewska", "maria@startup.io"),
        ("Piotr Zieliński", "piotr.z@consulting.pl"),
    ]
    subjects = [
        "Zapytanie ofertowe",
        "Re: Spotkanie w przyszłym tygodniu",
        "Nowa propozycja współpracy",
        "Faktura do opłacenia",
        "Zaproszenie na konferencję",
        "Pilne: aktualizacja umowy",
        "Podsumowanie projektu Q2",
        "Dostęp do systemu CRM",
        "Prośba o kontakt zwrotny",
        "Zamówienie #12345 - potwierdzenie",
    ]
    bodies = [
        "Dzień dobry,\n\nChciałbym zapytać o możliwość współpracy w zakresie dostarczenia rozwiązania CRM dla naszej firmy.\n\nPozdrawiam,\n{name}",
        "Cześć,\n\nPrzesyłam podsumowanie naszego ostatniego spotkania. Proszę o potwierdzenie ustaleń.\n\nDziękuję,\n{name}",
        "Szanowni Państwo,\n\nW załączeniu przesyłam naszą ofertę. Termin ważności: 30 dni.\n\nZ poważaniem,\n{name}",
        "Hej,\n\nCzy moglibyśmy umówić się na krótki call w tym tygodniu? Mam kilka pytań dotyczących integracji.\n\nPozdrawiam,\n{name}",
        "Witam,\n\nInformuję, że zaktualizowaliśmy warunki umowy. Proszę o zapoznanie się z załącznikiem.\n\n{name}",
    ]

    created = 0
    for i in range(count):
        idx = random.randint(0, len(senders) - 1)
        sender_name, sender_email = senders[idx]
        subject = random.choice(subjects)
        body = random.choice(bodies).format(name=sender_name)
        msg_id = f"<{uuid.uuid4()}@fake-imap>"
        date = email.utils.formatdate(time.time() - (count - i) * 3600, localtime=True)

        eml = (
            f"From: {sender_name} <{sender_email}>\r\n"
            f"To: crm@bizmopol.local\r\n"
            f"Subject: {subject}\r\n"
            f"Message-ID: {msg_id}\r\n"
            f"Date: {date}\r\n"
            f"MIME-Version: 1.0\r\n"
            f"Content-Type: text/plain; charset=UTF-8\r\n"
            f"\r\n"
            f"{body}\r\n"
        )

        fname = f"msg_{int(time.time())}_{i:04d}.eml"
        with open(os.path.join(MAILBOX_DIR, fname), "wb") as f:
            f.write(eml.encode("utf-8"))
        created += 1

    print(f"[fake-imap] Injected {created} fake emails into {MAILBOX_DIR}")
    return created


def load_messages():
    """Load messages and assign stable UIDs based on sorted filenames."""
    ensure_mailbox()
    messages = []
    uids = []
    files = sorted(os.listdir(MAILBOX_DIR))
    uid = 0
    for fname in files:
        if not fname.endswith(".eml"):
            continue
        uid += 1
        path = os.path.join(MAILBOX_DIR, fname)
        with open(path, "rb") as f:
            raw = f.read()
        messages.append(raw.decode("utf-8", errors="replace"))
        uids.append(uid)
    return messages, uids


class IMAPSession(threading.Thread):
    def __init__(self, conn, addr):
        super().__init__(daemon=True)
        self.conn = conn
        self.addr = addr
        self.selected = False
        self.messages = []
        self.uids = []

    def send(self, data):
        self.conn.sendall((data + "\r\n").encode())

    def run(self):
        try:
            self.handle()
        except (ConnectionResetError, BrokenPipeError):
            pass
        except Exception as e:
            print(f"[fake-imap] Session error: {e}")
        finally:
            self.conn.close()

    def handle(self):
        self.send("* OK Fake IMAP Server ready")
        buf = b""
        while True:
            chunk = self.conn.recv(4096)
            if not chunk:
                break
            buf += chunk
            while b"\r\n" in buf:
                line, buf = buf.split(b"\r\n", 1)
                self.process_line(line.decode("utf-8", errors="replace"))

    def process_line(self, line):
        parts = line.split(" ", 2)
        if len(parts) < 2:
            return
        tag = parts[0]
        cmd = parts[1].upper()
        args = parts[2] if len(parts) > 2 else ""

        if cmd == "CAPABILITY":
            self.send("* CAPABILITY IMAP4rev1 AUTH=PLAIN UIDPLUS IDLE")
            self.send(f"{tag} OK CAPABILITY completed")
        elif cmd == "LOGIN":
            self.send(f"{tag} OK LOGIN completed")
        elif cmd == "AUTHENTICATE":
            self.send(f"{tag} OK AUTHENTICATE completed")
        elif cmd == "SELECT":
            self.messages, self.uids = load_messages()
            self.selected = True
            self.send(f"* {len(self.messages)} EXISTS")
            self.send("* 0 RECENT")
            self.send("* FLAGS (\\Seen \\Answered \\Flagged \\Deleted \\Draft)")
            self.send(f"* OK [UIDNEXT {len(self.messages) + 1}]")
            self.send(f"* OK [UIDVALIDITY 1]")
            self.send(f"{tag} OK [READ-WRITE] SELECT completed")
        elif cmd == "UID" and args:
            self.handle_uid(tag, args)
        elif cmd == "SEARCH":
            self.handle_search(tag, args)
        elif cmd == "FETCH":
            self.handle_fetch(tag, args, use_uid=False)
        elif cmd == "NOOP":
            self.send(f"{tag} OK NOOP completed")
        elif cmd == "LOGOUT":
            self.send("* BYE Fake IMAP server logging out")
            self.send(f"{tag} OK LOGOUT completed")
        elif cmd == "LIST":
            self.send('* LIST (\\HasNoChildren) "/" "INBOX"')
            self.send(f"{tag} OK LIST completed")
        elif cmd == "NAMESPACE":
            self.send('* NAMESPACE (("" "/")) NIL NIL')
            self.send(f"{tag} OK NAMESPACE completed")
        elif cmd == "ID":
            self.send('* ID ("name" "FakeIMAP" "version" "1.0")')
            self.send(f"{tag} OK ID completed")
        else:
            self.send(f"{tag} BAD Command not recognized")

    def handle_uid(self, tag, args):
        parts = args.split(" ", 1)
        subcmd = parts[0].upper()
        rest = parts[1] if len(parts) > 1 else ""

        if subcmd == "FETCH":
            self.handle_fetch(tag, rest, use_uid=True)
        elif subcmd == "SEARCH":
            self.handle_uid_search(tag, rest)
        else:
            self.send(f"{tag} BAD UID command not supported")

    def handle_uid_search(self, tag, args):
        # Filter UIDs based on UID range if present (e.g. "UID 5:*")
        args_upper = args.upper()
        filtered = list(self.uids)

        if "UID" in args_upper:
            import re
            m = re.search(r'UID\s+(\d+):\*', args)
            if m:
                start_uid = int(m.group(1))
                filtered = [u for u in self.uids if u >= start_uid]

        if filtered:
            uid_list = " ".join(str(u) for u in filtered)
            self.send(f"* SEARCH {uid_list}")
        else:
            self.send("* SEARCH")
        self.send(f"{tag} OK UID SEARCH completed")

    def handle_search(self, tag, args):
        # SEARCH SINCE <date> or SEARCH ALL — return all sequence numbers
        seq_nums = list(range(1, len(self.messages) + 1))
        if seq_nums:
            self.send(f"* SEARCH {' '.join(str(s) for s in seq_nums)}")
        else:
            self.send("* SEARCH")
        self.send(f"{tag} OK SEARCH completed")

    def handle_fetch(self, tag, args, use_uid=False):
        seq_end = args.find(" ")
        if seq_end < 0:
            self.send(f"{tag} BAD FETCH syntax error")
            return
        seq_str = args[:seq_end]
        items_str = args[seq_end+1:].strip()
        items_upper = items_str.upper()

        indices = self.parse_sequence(seq_str, use_uid)

        for idx in indices:
            if idx < 0 or idx >= len(self.messages):
                continue
            msg = self.messages[idx]
            uid = self.uids[idx]
            seq_num = idx + 1

            # Split message into header and body
            if "\r\n\r\n" in msg:
                header_part, body_part = msg.split("\r\n\r\n", 1)
            else:
                header_part, body_part = msg, ""
            header_part += "\r\n"

            response_parts = []
            literals = []  # (label, data) pairs to send as literals

            if "FLAGS" in items_upper:
                response_parts.append("FLAGS (\\Seen)")
            if "UID" in items_upper:
                response_parts.append(f"UID {uid}")

            # Build segments: list of (text_part, literal_bytes_or_None)
            import re
            segments = []  # list of (prefix_text, literal_data|None)

            if "BODY[HEADER.FIELDS" in items_upper:
                m = re.search(r'HEADER\.FIELDS\s*\(([^)]+)\)', items_str, re.IGNORECASE)
                if m:
                    wanted = [h.strip().upper() for h in m.group(1).split()]
                    filtered_lines = []
                    for hline in header_part.split("\r\n"):
                        if ":" in hline:
                            key = hline.split(":")[0].strip().upper()
                            if key in wanted:
                                filtered_lines.append(hline)
                    filtered_header = "\r\n".join(filtered_lines) + "\r\n\r\n"
                else:
                    filtered_header = header_part + "\r\n"

                hdr_bytes = filtered_header.encode("utf-8")
                label = re.search(r'(BODY\[HEADER\.FIELDS\s*\([^)]+\)\])', items_str, re.IGNORECASE)
                label_str = label.group(1) if label else "BODY[HEADER]"
                segments.append((label_str, hdr_bytes))

            elif "BODY[]" in items_upper or "BODY.PEEK[]" in items_upper or "RFC822" in items_upper:
                msg_bytes = msg.encode("utf-8")
                segments.append(("BODY[]", msg_bytes))

            elif "BODY[HEADER]" in items_upper or "BODY.PEEK[HEADER]" in items_upper:
                hdr_bytes = (header_part + "\r\n").encode("utf-8")
                segments.append(("BODY[HEADER]", hdr_bytes))

            if "BODY[TEXT]" in items_upper or "BODY.PEEK[TEXT]" in items_upper:
                body_bytes = body_part.encode("utf-8")
                segments.append(("BODY[TEXT]", body_bytes))

            # Send IMAP FETCH response with proper literal format:
            # {N}\r\n followed by N bytes, then \r\n before next part
            if segments:
                line = f"* {seq_num} FETCH ({' '.join(response_parts)}"
                for i, (label, data) in enumerate(segments):
                    line += f" {label} {{{len(data)}}}"
                    self.conn.sendall((line + "\r\n").encode())
                    self.conn.sendall(data)
                    self.conn.sendall(b"\r\n")
                    if i < len(segments) - 1:
                        line = ""
                    else:
                        self.send(")")
            elif response_parts:
                self.send(f"* {seq_num} FETCH ({' '.join(response_parts)})")

        self.send(f"{tag} OK FETCH completed")

    def parse_sequence(self, seq_str, use_uid):
        """Parse IMAP sequence set like '1:*', '1,2,3', '5:10'"""
        indices = []
        max_val = len(self.messages)
        if max_val == 0:
            return []

        for part in seq_str.split(","):
            if ":" in part:
                start_s, end_s = part.split(":", 1)
                start = max_val if start_s == "*" else int(start_s)
                end = max_val if end_s == "*" else int(end_s)
                if start > end:
                    start, end = end, start
                for v in range(start, end + 1):
                    if use_uid:
                        if v in self.uids:
                            indices.append(self.uids.index(v))
                    else:
                        idx = v - 1
                        if 0 <= idx < max_val:
                            indices.append(idx)
            else:
                v = max_val if part == "*" else int(part)
                if use_uid:
                    if v in self.uids:
                        indices.append(self.uids.index(v))
                else:
                    idx = v - 1
                    if 0 <= idx < max_val:
                        indices.append(idx)

        return sorted(set(indices))


def run_control_server(port):
    """Simple TCP control server to inject emails at runtime."""
    srv = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    srv.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    srv.bind(("0.0.0.0", port))
    srv.listen(1)
    print(f"[fake-imap] Control server on port {port} (send 'inject N')")
    while True:
        try:
            conn, _ = srv.accept()
            data = conn.recv(1024).decode().strip()
            if data.startswith("inject"):
                parts = data.split()
                n = int(parts[1]) if len(parts) > 1 else 1
                inject_fake_emails(n)
                conn.sendall(f"OK injected {n}\n".encode())
            conn.close()
        except Exception:
            pass


def run_auto_generator(min_sec, max_sec):
    """Periodically inject a single fake email at random intervals."""
    print(f"[fake-imap] Auto-generator: new email every {min_sec}-{max_sec}s")
    while True:
        delay = random.uniform(min_sec, max_sec)
        time.sleep(delay)
        inject_fake_emails(1)


def main():
    parser = argparse.ArgumentParser(description="Fake IMAP server for testing")
    parser.add_argument("--port", type=int, default=1143, help="IMAP port")
    parser.add_argument("--inject", type=int, default=5, help="Inject N fake emails on startup")
    parser.add_argument("--auto-generate", action="store_true", help="Auto-generate emails periodically")
    parser.add_argument("--auto-min", type=int, default=10, help="Min seconds between auto-generated emails")
    parser.add_argument("--auto-max", type=int, default=15, help="Max seconds between auto-generated emails")
    args = parser.parse_args()

    if args.inject > 0:
        inject_fake_emails(args.inject)

    # Control server thread
    ctl_thread = threading.Thread(target=run_control_server, args=(args.port + CONTROL_PORT_OFFSET,), daemon=True)
    ctl_thread.start()

    if args.auto_generate:
        gen_thread = threading.Thread(target=run_auto_generator, args=(args.auto_min, args.auto_max), daemon=True)
        gen_thread.start()

    # IMAP server
    srv = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    srv.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    srv.bind(("0.0.0.0", args.port))
    srv.listen(5)
    print(f"[fake-imap] IMAP server listening on port {args.port}")
    print(f"[fake-imap] Accepts any login credentials")
    print(f"[fake-imap] Mailbox dir: {MAILBOX_DIR}")

    def shutdown(sig, frame):
        print("\n[fake-imap] Shutting down...")
        srv.close()
        sys.exit(0)
    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)

    while True:
        try:
            conn, addr = srv.accept()
            print(f"[fake-imap] Connection from {addr}")
            session = IMAPSession(conn, addr)
            session.start()
        except OSError:
            break


if __name__ == "__main__":
    main()
