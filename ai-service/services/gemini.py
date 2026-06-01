import time
import os
import requests
import random
from dotenv import load_dotenv

load_dotenv()

def generate_content(prompt: str, max_retries: int = 3) -> str:
    groq_keys = os.environ.get("GROQ_API_KEYS", os.environ.get("GROQ_API_KEY", ""))
    gemini_keys = os.environ.get("GEMINI_API_KEYS", os.environ.get("GEMINI_API_KEY", ""))
    
    groq_api_key = random.choice([k.strip() for k in groq_keys.split(",") if k.strip()]) if groq_keys else None
    gemini_api_key = random.choice([k.strip() for k in gemini_keys.split(",") if k.strip()]) if gemini_keys else None
    
    if groq_api_key:
        url = "https://api.groq.com/openai/v1/chat/completions"
        headers = {
            "Authorization": f"Bearer {groq_api_key}",
            "Content-Type": "application/json"
        }
        payload = {
            "model": "llama-3.1-8b-instant",
            "messages": [{"role": "user", "content": prompt}]
        }
        
        for attempt in range(max_retries):
            response = requests.post(url, headers=headers, json=payload)
            if response.status_code == 200:
                try:
                    return response.json()["choices"][0]["message"]["content"]
                except (KeyError, IndexError) as e:
                    raise Exception(f"Unexpected response format from Groq: {response.json()}")
            elif response.status_code == 503 or response.status_code == 429:
                if attempt < max_retries - 1:
                    time.sleep(2 ** attempt)
                    continue
                else:
                    raise Exception(f"Groq API Error {response.status_code}: High demand. Please try again later.")
            else:
                raise Exception(f"Groq API Error {response.status_code}: {response.text}")

    elif gemini_api_key:
        url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key={gemini_api_key}"
        headers = {"Content-Type": "application/json"}
        payload = {
            "contents": [
                {
                    "parts": [{"text": prompt}]
                }
            ]
        }
        
        for attempt in range(max_retries):
            response = requests.post(url, headers=headers, json=payload)
            if response.status_code == 200:
                data = response.json()
                try:
                    return data["candidates"][0]["content"]["parts"][0]["text"]
                except (KeyError, IndexError) as e:
                    raise Exception(f"Unexpected response format from Gemini: {data}")
            elif response.status_code == 503 or response.status_code == 429:
                if attempt < max_retries - 1:
                    time.sleep(2 ** attempt) # Exponential backoff
                    continue
                else:
                    raise Exception(f"Gemini API Error {response.status_code}: High demand. Please try again later.")
            else:
                raise Exception(f"Gemini API Error {response.status_code}: {response.text}")
    else:
        raise ValueError("Neither GROQ_API_KEY nor GEMINI_API_KEY environment variables are set.")


def generate_hint(problem_statement: str, user_code: str, editorial_code: str) -> str:
    prompt = f"""
    You are an expert programming interviewer and mentor. The user is stuck on a competitive programming problem.
    Your goal is to give a subtle hint about what is wrong with their approach or code.
    
    CRITICAL RULES: 
    1. DO NOT GIVE THEM THE FULL SOLUTION OR CODE. Just point them in the right direction.
    2. Format your response STRICTLY as a bulleted list using the '-' character. 
    3. Provide exactly 2-3 short, actionable bullet points. Do not include conversational filler like "Here is a hint".

    Problem Statement:
    {problem_statement}

    Editorial Code:
    {editorial_code}

    User Code:
    {user_code}
    """
    
    return generate_content(prompt)

def generate_feedback(problem_statement: str, user_code: str, editorial_code: str) -> str:
    prompt = f"""
    You are an expert code reviewer. The user's code was ACCEPTED, but they want to know how to optimize it.
    
    CRITICAL RULES:
    1. Compare the user's code to the editorial code.
    2. Format your response STRICTLY as a bulleted list using the '-' character.
    3. Do not include conversational filler like "Here is your feedback".
    4. You MUST provide exactly three bullet points in this exact structure:
       - Time Complexity: (Analyze their TC vs optimal)
       - Space Complexity: (Analyze their SC vs optimal)
       - Suggestion: (One concrete way to improve cleanliness or performance)

    Problem Statement:
    {problem_statement}

    Editorial Code:
    {editorial_code}

    User Code:
    {user_code}
    """
    
    return generate_content(prompt)
