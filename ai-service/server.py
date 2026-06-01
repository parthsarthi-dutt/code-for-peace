import grpc
from concurrent import futures
import time
import json

import sys
sys.path.append('./proto')

import evaluation_pb2
import evaluation_pb2_grpc
from services.gemini import generate_hint, generate_feedback
from services.interview import generate_first_question, process_response

class EvaluationServiceServicer(evaluation_pb2_grpc.EvaluationServiceServicer):
    def GenerateHint(self, request, context):
        try:
            print("Received GenerateHint request")
            hint_text = generate_hint(request.problem_statement, request.user_code, request.editorial_code)
            return evaluation_pb2.HintResponse(hint=hint_text)
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return evaluation_pb2.HintResponse()

    def GenerateFeedback(self, request, context):
        try:
            print("Received GenerateFeedback request")
            feedback_text = generate_feedback(request.problem_statement, request.user_code, request.editorial_code)
            return evaluation_pb2.FeedbackResponse(feedback=feedback_text)
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return evaluation_pb2.FeedbackResponse()

    def StartInterviewSession(self, request, context):
        try:
            print(f"Starting interview: level={request.level}, duration={request.duration}")
            question_text, audio_bytes = generate_first_question(request.level, request.duration)
            return evaluation_pb2.InterviewQuestionResponse(
                question_text=question_text,
                audio_bytes=audio_bytes,
                user_transcript="",
            )
        except Exception as e:
            print(f"StartInterviewSession error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return evaluation_pb2.InterviewQuestionResponse()

    def ProcessInterviewResponse(self, request, context):
        try:
            print(f"Processing interview response: level={request.level}")
            chat_history = json.loads(request.chat_history_json) if request.chat_history_json else []
            next_question, audio_bytes, transcript = process_response(
                request.level,
                request.duration,
                chat_history,
                request.audio_bytes,
                request.time_up,
                request.system_action,
                request.code,
            )
            return evaluation_pb2.InterviewQuestionResponse(
                question_text=next_question,
                audio_bytes=audio_bytes,
                user_transcript=transcript,
            )
        except Exception as e:
            print(f"ProcessInterviewResponse error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return evaluation_pb2.InterviewQuestionResponse()

def serve():
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=10),
        options=[
            ('grpc.max_send_message_length', 50 * 1024 * 1024),
            ('grpc.max_receive_message_length', 50 * 1024 * 1024),
        ]
    )
    evaluation_pb2_grpc.add_EvaluationServiceServicer_to_server(EvaluationServiceServicer(), server)
    server.add_insecure_port('[::]:50051')
    print("AI Evaluation Service started on port 50051...")
    server.start()
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()
