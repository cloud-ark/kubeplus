from flask import Flask, request, jsonify
from ollama import generate
import os

app = Flask(__name__)
MODEL = os.getenv("MODEL_NAME", "gemma:2b")

@app.get("/healthz")
def healthz():
    return "ok", 200

@app.route("/crailogs", methods=["POST"])
def cr_ai_logs():
	data = request.get_json(force=True)
	logs = data.get("logs", "")
	prompt = """
				You are a Kubernetes SRE copilot. Analyze the raw logs below and produce a strictly formatted JSON report.

				TASKS:
					1. Provide a 1-2 line "overall_status" summarizing the observed issues.
					2. Detect incidents (SEV1-SEV3), each including:
						- pods
						- patterns
						- sample_log
						- likely_root_cause
						- impact
						- recommended_actions
					3. Output ONLY the following JSON structure:

				{
					"title": "<string>",
					"incidents": [
						{
							"pods": ["<pod>", "..."],
							"patterns": ["<pattern>", "..."],
							"sample_log": "<string>",
							"likely_root_cause": "<string>",
							"impact": "<string>",
							"recommended_actions": ["<action>", "..."]
						}
					]
				}

				RULES:
					- Consider log lines as incidents if they contain keywords: ERROR, FAIL, panic, CrashLoopBackOff, OOMKilled, exception.
					- For each detected incident, fill the "incidents" array with all required fields.
					- Be concise; overall_status should be 1-2 lines.
					- Stick strictly to the JSON structure; DO NOT add any fields or text outside the JSON.
					- Do not be too verbose. Be concise and stick exactly to the JSON format.

				LOGS:
			""".strip()
	prompt = f"{prompt}\n{logs}"
	try:
		response = generate(model=f"{MODEL}", prompt=prompt)
		return jsonify({"output": response.get("response", "")}), 200
	except Exception as e:
		return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=int(os.getenv("PORT", "8080")))
