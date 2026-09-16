"""Render the limited Markdown used in legal/*.md into static HTML pages."""
from pathlib import Path
import html
import re

ROOT = Path(__file__).resolve().parents[1]
STYLE = '''
:root { color-scheme: light; --ink:#13231d; --muted:#65716c; --green:#07563f; }
* { box-sizing:border-box; } body { margin:0; background:#f6f3ea; color:var(--ink); font-family:system-ui,sans-serif; line-height:1.8; }
main { max-width:880px; margin:auto; padding:32px 20px 56px; } header, footer nav { display:flex; flex-wrap:wrap; gap:18px; }
a { color:var(--green); overflow-wrap:anywhere; } h1 { font-size:clamp(28px,5vw,40px); line-height:1.3; } h2 { margin-top:32px; font-size:23px; }
section, article { background:white; padding:24px; border:1px solid #d9dfd8; border-radius:18px; margin:24px 0; overflow-wrap:anywhere; }
blockquote { margin:16px 0; padding:12px; background:#fff5ce; } footer { margin-top:32px; font-size:14px; } label { display:block; margin:20px 0; } input { width:20px; height:20px; vertical-align:middle; }
button, .button { display:inline-block; padding:12px 18px; background:var(--green); color:white; border:0; border-radius:8px; font:inherit; cursor:pointer; text-decoration:none; } button:disabled { background:#737c76; cursor:not-allowed; }
:focus-visible { outline:3px solid #0c7c59; outline-offset:4px; } .note { color:var(--muted); font-size:14px; }
@media(max-width:480px) { section,article { padding:18px; } }
'''
NAV = '<nav aria-label="関連ページ"><a href="/contact">お問い合わせ</a><a href="/terms">利用規約</a><a href="/privacy">プライバシーポリシー</a></nav>'

def page(title, body, draft=False):
    robots = '<meta name="robots" content="noindex">' if draft else ''
    return f'''<!doctype html>
<html lang="ja"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="referrer" content="no-referrer">{robots}<title>{title} | QR Studio</title><style>{STYLE}</style></head>
<body><main><header><a href="/">QR Studio</a><a href="/docs">APIドキュメント</a></header>{body}<footer>{NAV}</footer></main></body></html>\n'''

def inline(s):
    s = html.escape(s)
    return re.sub(r'\[([^\]]+)\]\((https?://[^\s)]+|/[^\s)]*)\)', r'<a href="\2">\1</a>', s)

def render(source):
    blocks = []
    for block in source.strip().split('\n\n'):
        if block.startswith('#'):
            marks, text = block.split(' ', 1)
            blocks.append(f'<h{len(marks)}>{inline(text)}</h{len(marks)}>')
        elif block.startswith('- '):
            blocks.append('<ul>' + ''.join('<li>' + inline(line[2:]) + '</li>' for line in block.splitlines()) + '</ul>')
        elif block.startswith('> '):
            blocks.append('<blockquote>' + inline(block[2:]) + '</blockquote>')
        else:
            blocks.append('<p>' + '<br>'.join(inline(line.strip()) for line in block.splitlines()) + '</p>')
    return '\n'.join(blocks)

if __name__ == '__main__':
    for name, title in [('terms', '利用規約'), ('privacy', 'プライバシーポリシー')]:
        source = (ROOT / 'legal' / f'{name}.md').read_text(encoding="utf-8")
        (ROOT / f'{name}.html').write_text(page(title, '<article>' + render(source) + '</article>', '（草案）' in source), encoding="utf-8")
    contact = (ROOT / 'legal' / 'contact-body.html').read_text(encoding="utf-8")
    (ROOT / 'contact.html').write_text(page('お問い合わせ', contact), encoding="utf-8")
