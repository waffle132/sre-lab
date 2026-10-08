from urllib.request import Request, urlopen
import json
import sys
from urllib.error import URLError

payload = json.dumps({"URL": "http://example.com"}).encode('utf-8')  # 将字典转换为JSON并编码为字节
request = Request("http://localhost:8080/shorten", 
                  data=payload,
                  headers={"Content-Type": "application/json"},
                  method="POST")
#组装request
try:
    with urlopen(request,timeout=3) as response:
        status = response.status
        response_text = response.read().decode('utf-8')
except (URLError, TimeoutError) as e:
    print(f"failed: {e}")
    sys.exit(1)
#测试urlopen
if status != 200:
    print(f"failed: {status}")
    sys.exit(1)

try:
    result = json.loads(response_text)
except json.JSONDecodeError as e:
    print(f"failed: {e}")
    sys.exit(1)
#测试json格式合法
if not isinstance(result, dict):
    print(f"failed: result is not a dict")
    sys.exit(1)
#判断result为字典    
code = result.get("code")

if not isinstance(code, str) or not code.strip():
    print(f"failed: code is not a valid string")
    sys.exit(1)
print(f"success: code = {code}")
