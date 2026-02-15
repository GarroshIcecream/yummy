import sys
import json

from recipe_scrapers import scrape_me

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 fetch_recipe_json.py <url>", file=sys.stderr)
        sys.exit(1)
    url = sys.argv[1]
    try:
        scraper = scrape_me(url)
        print(json.dumps(scraper.to_json(), ensure_ascii=False, indent=2))
    except Exception as e:
        print(str(e), file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
