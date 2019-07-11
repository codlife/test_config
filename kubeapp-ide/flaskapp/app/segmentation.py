import jieba
from flask import Flask, abort, request, jsonify

app = Flask(__name__)

@app.route('/segmentation/0.0.1', methods=['POST'])
def segmentation():
    if not request.json or 'sentence' not in request.json:
        print(request.json)
    sentence = request.json['sentence']
    segments = jieba.cut(sentence)
    words = list()
    for s in segments:
        words.append(s)
    return jsonify({"result":"/".join(words)})
@app.route('/index', methods=['GET'])
def index():
	return jsonify({"result": "you are right"})

@app.route('/segmentation/index', methods=['GET'])
def hello():
    return "hello world"


if __name__ == '__main__':
    app.run(host="0.0.0.0", port=8080, debug=True)
