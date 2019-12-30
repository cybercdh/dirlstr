import requests

phishtank_url = "http://data.phishtank.com/data/online-valid.json"
headers = {
  'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36'
}
try:  
  r = requests.get(phishtank_url, allow_redirects=True, timeout=5, stream=True, headers=headers)
except requests.exceptions.RequestException:
  sys.exit()

parsed_json = r.json()
  # go phishing baby!
for entry in parsed_json:
  url = entry['url'].strip()
  print (url)