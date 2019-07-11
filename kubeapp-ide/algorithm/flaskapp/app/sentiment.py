from snownlp import SnowNLP
from flask import Flask, abort, request, jsonify

app = Flask(__name__)

@app.route('/sentiment/0.0.1', methods=['POST'])
def sentiment():
    if not request.json or 'document' not in request.json:
        print(request.json)
    document = request.json['document']
    res = SnowNLP(document)
    return jsonify({"result": str(res.sentiments)})
@app.route('/info', methods=['GET'])
def index():
	return jsonify({"result": "you are right"})

@app.route('/sentiment/info', methods=['GET'])
def info():
    return "Algorithm Description: determine positive or negative sentiment from text.\n"


if __name__ == '__main__':
    app.run(host="0.0.0.0", port=8080, debug=True)
