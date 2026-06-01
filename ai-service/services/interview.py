from __future__ import annotations

"""
AI Interview Service
- Uses Gemini/Groq LLM for question generation & evaluation
- Uses Groq Whisper for Speech-to-Text
- Uses Piper TTS (Docker container on port 5002) for Text-to-Speech
"""

import os
import json
import time
import requests
from dotenv import load_dotenv
from urllib.parse import quote_plus

load_dotenv()

import io
from gtts import gTTS

# Use gTTS for reliable Text-to-Speech without an external container

# ─── Topic pools per difficulty ─────────────────────
TOPIC_POOLS = {
    "easy": [
        "Arrays and basic operations",
        "String manipulation",
        "Basic sorting algorithms",
        "Hash maps and frequency counting",
        "Two pointers technique",
        "Basic recursion",
        "Stack and Queue operations",
        "Linear search and Binary search",
        "Time and Space complexity analysis",
        "Linked list basics",
    ],
    "medium": [
        "Binary Search on answer",
        "Sliding window technique",
        "Backtracking and pruning",
        "Graph BFS and DFS",
        "Dynamic Programming (1D)",
        "Greedy algorithms",
        "Tree traversals and properties",
        "Prefix sums and difference arrays",
        "Topological sorting",
        "Union-Find / Disjoint Set Union",
    ],
    "hard": [
        "Dynamic Programming (2D and bitmask DP)",
        "Segment Trees and Fenwick Trees",
        "Network Flow and matching",
        "Advanced graph algorithms (Dijkstra, Floyd-Warshall)",
        "Trie data structures",
        "Suffix arrays and string hashing",
        "Heavy-Light Decomposition",
        "Centroid Decomposition",
        "Convex Hull Trick",
        "Matrix Exponentiation",
    ],
}


def _get_api_keys(prefix: str) -> list[str]:
    keys = []
    
    # Check exact prefix first e.g. GROQ_API_KEY
    exact_key = os.environ.get(prefix, "")
    if exact_key:
        keys.extend([k.strip() for k in exact_key.split(",") if k.strip()])
        
    # Check plural comma separated e.g. GROQ_API_KEYS
    multi_key = os.environ.get(f"{prefix}S", "")
    if multi_key:
        keys.extend([k.strip() for k in multi_key.split(",") if k.strip()])
        
    base_key = os.environ.get(prefix)
    if base_key and base_key not in keys:
        keys.append(base_key)
        
    for i in range(1, 10):
        k = os.environ.get(f"{prefix}_{i}")
        if k and k not in keys:
            keys.append(k)
            
    return list(dict.fromkeys(keys))

def _call_llm(prompt: str) -> str:
    """Call Groq or Gemini LLM with the given prompt, using key rotation."""
    groq_keys = _get_api_keys("GROQ_API_KEY")
    gemini_keys = _get_api_keys("GEMINI_API_KEY")

    # Cycle through Groq keys
    for api_key in groq_keys:
        url = "https://api.groq.com/openai/v1/chat/completions"
        headers = {
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        }
        payload = {
            "model": "llama-3.1-8b-instant",
            "messages": [{"role": "user", "content": prompt}],
        }
        try:
            response = requests.post(url, headers=headers, json=payload, timeout=10)
            if response.status_code == 200:
                return response.json()["choices"][0]["message"]["content"]
            print(f"Groq error ({response.status_code}): {response.text}")
        except Exception:
            pass # Move to next key

    # Cycle through Gemini keys if Groq fails
    for api_key in gemini_keys:
        url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key={api_key}"
        headers = {"Content-Type": "application/json"}
        payload = {"contents": [{"parts": [{"text": prompt}]}]}
        try:
            response = requests.post(url, headers=headers, json=payload, timeout=10)
            if response.status_code == 200:
                data = response.json()
                return data["candidates"][0]["content"]["parts"][0]["text"]
            print(f"Gemini error ({response.status_code}): {response.text}")
        except Exception:
            pass

    raise Exception("All Groq and Gemini API keys failed or were rate limited.")


def _transcribe_audio(audio_bytes: bytes) -> str:
    """Use Groq Whisper API for Speech-to-Text with key rotation."""
    groq_keys = _get_api_keys("GROQ_API_KEY")
    if not groq_keys:
        raise ValueError("GROQ_API_KEY required for speech-to-text")

    for api_key in groq_keys:
        url = "https://api.groq.com/openai/v1/audio/transcriptions"
        headers = {"Authorization": f"Bearer {api_key}"}
        files = {"file": ("audio.webm", audio_bytes, "audio/webm")}
        data = {"model": "whisper-large-v3", "language": "en"}
        try:
            response = requests.post(url, headers=headers, files=files, data=data, timeout=15)
            if response.status_code == 200:
                return response.json().get("text", "")
            print(f"Whisper STT error ({response.status_code}): {response.text}")
        except Exception as e:
            print(f"Whisper request failed: {e}")
            pass

    raise Exception("All Groq Whisper API keys failed.")


def _synthesize_speech(text: str) -> bytes:
    """Use gTTS for reliable Text-to-Speech."""
    try:
        tts = gTTS(text=text, lang='en', tld='com')
        fp = io.BytesIO()
        tts.write_to_fp(fp)
        return fp.getvalue()
    except Exception as e:
        print(f"gTTS error: {e}")
        return b""


def generate_first_question(level: str, duration: int) -> tuple[str, bytes]:
    """
    Generate the opening question for an AI interview session.
    Returns (question_text, audio_bytes).
    """
    topics = TOPIC_POOLS.get(level, TOPIC_POOLS["easy"])
    topic_list = ", ".join(topics[:5])

    prompt = f"""You are a senior software engineer conducting a technical coding interview.
The interview is {duration} minutes long at the "{level}" difficulty level.

Generate your FIRST interview question. Follow these rules:
1. Start with a brief, warm greeting (1 sentence max).
2. Ask the candidate to briefly introduce themselves and their background.
3. Do NOT ask any technical or coding questions yet.
4. Keep the entire response under 50 words.
5. Do NOT use markdown formatting or bullet points. Speak naturally as an interviewer would.
6. Do NOT use placeholders like [Candidate Name]! Just say "Hi there" or use a generic friendly greeting."""

    question_text = _call_llm(prompt)
    audio_bytes = _synthesize_speech(question_text)
    return question_text, audio_bytes


def process_response(level: str, duration: int, chat_history: list, audio_bytes: bytes, time_up: bool = False, system_action: str = "") -> tuple[str, bytes, str]:
    """
    Process user's audio response or system actions and generate the next interviewer question.
    Returns (next_question_text, audio_bytes, user_transcript).
    """
    # Step 1: Transcribe user audio (if no system action)
    if system_action:
        user_transcript = "[Candidate remained silent]"
    else:
        user_transcript = _transcribe_audio(audio_bytes)
        if not user_transcript.strip():
            fallback = "I didn't catch that. Could you please repeat your answer?"
            fallback_audio = _synthesize_speech(fallback)
            return fallback, fallback_audio, ""

    # Step 2: Build conversation context
    conversation = ""
    for entry in chat_history:
        role = entry.get("role", "unknown")
        text = entry.get("text", "")
        if role == "interviewer":
            conversation += f"Interviewer: {text}\n"
        else:
            conversation += f"Candidate: {text}\n"

    conversation += f"Candidate: {user_transcript}\n"

    topics = TOPIC_POOLS.get(level, TOPIC_POOLS["easy"])
    topic_list = ", ".join(topics)

    if system_action == "idle_nudge":
        prompt = f"""You are a senior software engineer conducting a technical coding interview.
The candidate has been silent for 1 minute.
Politely ask them if they are still thinking, or if they need a hint. Keep it under 20 words. Do not use markdown."""
    elif system_action == "idle_skip":
        prompt = f"""You are a senior software engineer conducting a technical coding interview.
The candidate has remained silent even after being prompted. 
Assume they don't know the answer. Briefly state that we will move on, and ask a NEW algorithmic question. Keep it under 60 words. Do not use markdown."""
    elif time_up:
        prompt = f"""You are a senior software engineer conducting a technical coding interview.
The interview time has now expired.

Here is the conversation so far:
{conversation}

Based on the candidate's last response, give a very brief suggestion or feedback on it (1-2 sentences), and then formally conclude the interview by thanking them for their time. 
Do not ask any new questions.
Keep your response under 80 words total. Be natural and conversational. Do not use markdown, bullet points, or code blocks."""
    else:
        prompt = f"""You are a senior software engineer conducting a technical coding interview.
The interview is {duration} minutes long at the "{level}" difficulty level.
Topics you can cover: {topic_list}

Here is the conversation so far:
{conversation}

If the candidate just introduced themselves, acknowledge it briefly and then ask your FIRST algorithmic coding question based on the topics above.
Otherwise, based on the candidate's last response, do ONE of the following:
- If their answer is correct/good: Briefly acknowledge it (1 sentence), then ask a NEW algorithmic question or dive deeper.
- If their answer is partially correct: Point out what's missing and ask a follow-up to guide them.
- If their answer is wrong: Gently correct them with a brief explanation and move to the next question.

Important Focus Areas:
- Focus primarily on the core algorithm and problem-solving approach.
- Do NOT focus heavily on edge cases. You may ask about them occasionally, but keep the focus on the main algorithmic logic.

Rules:
1. Keep your response under 80 words total.
2. Be conversational and natural—you are speaking, not writing.
3. Do NOT use markdown, bullet points, or code blocks.
4. Do NOT say "Let's move on to" or use robotic transitions.
5. Ask only ONE question at a time."""

    next_question = _call_llm(prompt)
    next_audio = _synthesize_speech(next_question)
    return next_question, next_audio, user_transcript
