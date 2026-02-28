#!/usr/bin/env python3
"""
Generate articles.md from the Dev Leader Google Search Console MCP tag page.
Usage: python scripts/generate-articles.py docs/articles.md
"""

import sys

PLACEHOLDER = """# Articles & Blog Posts

Blog posts and articles about this MCP server from [Dev Leader](https://www.devleader.ca).

!!! tip "Stay Updated"
    Follow [Dev Leader](https://www.devleader.ca) for articles about MCP servers, Google Search Console, and C#/.NET development.

*No articles published yet. Check back soon!*
"""


def main():
    if len(sys.argv) < 2:
        print('Usage: python generate-articles.py <output-file>')
        sys.exit(1)

    output_file = sys.argv[1]

    with open(output_file, 'w', encoding='utf-8') as f:
        f.write(PLACEHOLDER)

    print(f'Generated {output_file}')


if __name__ == "__main__":
    main()
