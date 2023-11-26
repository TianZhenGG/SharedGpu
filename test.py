import requests

# https://openai-proxy-api.pages.dev/api
# https://openai.451024.xyz
# https://openai-proxy-api.pages.dev/api
    
class chat_bot():
    def __init__(self):
        self.url = "https://openai-proxy-api.pages.dev/api/v1/chat/completions"
        self.api_key = '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'
        self.headers = {
          'Authorization': f'Bearer {self.api_key}',
          'Content-Type': 'application/json'
        }
        self.payload = {
          "model": "gpt-4",
          "messages": [
            {
              "role": "user",
              # "content": "你是谁？"
            }
          ]
        }

    def chat(self, content):
        self.payload['messages'][0]['content'] = content
        try:
            response = requests.request("POST", self.url, headers=self.headers, json=self.payload)
            response.raise_for_status()
            return response.json()['choices'][0]['message']['content']
        except requests.exceptions.HTTPError as err:
            raise SystemExit(err)
        

send = chat_bot()
while True:
    content = "111"
    print(send.chat(content))

