from __future__ import annotations

"""
AI Interview Service
- Uses Gemini/Groq LLM for question generation & evaluation
- Uses Groq Whisper for Speech-to-Text
- Uses gTTS for Text-to-Speech
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

# ─── Topic pools per difficulty ─────────────────────
# Easy  → TCS / Infosys / Wipro / Accenture / entry-level product companies
# Medium → Amazon / Microsoft / Flipkart / Adobe / Atlassian / Uber (early rounds)
# Hard  → Google / Meta / Rubrik / Snowflake / Stripe / Jane Street / competitive-level

TOPIC_POOLS = {
    "easy": [
        # Core DS / Algo fundamentals tested at TCS, Infosys, Wipro, Cognizant
        {
            "topic": "Arrays and basic operations",
            "example_questions": [
                "Find the second largest element in an unsorted array.",
                "Rotate an array to the right by k steps.",
                "Check if an array is a palindrome.",
            ],
            "companies": ["TCS", "Infosys", "Wipro"],
        },
        {
            "topic": "String manipulation",
            "example_questions": [
                "Reverse words in a sentence without reversing each word.",
                "Check if two strings are anagrams.",
                "Find the first non-repeating character in a string.",
            ],
            "companies": ["Accenture", "Cognizant", "Capgemini"],
        },
        {
            "topic": "Hash maps and frequency counting",
            "example_questions": [
                "Find all elements that appear more than n/3 times.",
                "Group words that are anagrams of each other.",
                "Two-sum problem using a hash map.",
            ],
            "companies": ["TCS Digital", "Hexaware", "Mphasis"],
        },
        {
            "topic": "Basic sorting algorithms",
            "example_questions": [
                "Sort an array of 0s, 1s, and 2s without using extra space.",
                "Find the kth largest element using partial sort.",
            ],
            "companies": ["TCS", "Infosys", "Wipro"],
        },
        {
            "topic": "Two pointers technique",
            "example_questions": [
                "Find a pair in a sorted array that sums to a target.",
                "Remove duplicates from a sorted array in-place.",
                "Move all zeroes to the end of the array.",
            ],
            "companies": ["Zoho", "Freshworks", "EPAM"],
        },
        {
            "topic": "Basic recursion and mathematical problems",
            "example_questions": [
                "Compute power(x, n) using fast exponentiation.",
                "Find GCD of two numbers using Euclid's algorithm.",
                "Check if a number is prime with optimal complexity.",
            ],
            "companies": ["TCS", "Infosys", "Persistent"],
        },
        {
            "topic": "Stack and Queue operations",
            "example_questions": [
                "Check if a string of brackets is balanced.",
                "Implement a queue using two stacks.",
                "Design a stack that supports getMin in O(1).",
            ],
            "companies": ["Accenture", "L&T Infotech", "Mindtree"],
        },
        {
            "topic": "Binary search fundamentals",
            "example_questions": [
                "Find the first and last position of a target in a sorted array.",
                "Search in a rotated sorted array.",
            ],
            "companies": ["TCS Digital", "Zoho", "Freshworks"],
        },
        {
            "topic": "Linked list basics",
            "example_questions": [
                "Detect a cycle in a linked list.",
                "Reverse a linked list both iteratively and recursively.",
                "Find the middle element of a linked list in one pass.",
            ],
            "companies": ["Wipro Elite", "HCL", "Tech Mahindra"],
        },
        {
            "topic": "Time and space complexity analysis",
            "example_questions": [
                "Analyze the complexity of bubble sort vs merge sort.",
                "Why is hash map lookup O(1) amortized but not worst-case?",
            ],
            "companies": ["All entry-level companies"],
        },
    ],

    "medium": [
        # Patterns tested at Amazon, Microsoft, Adobe, Flipkart, Atlassian, Swiggy
        {
            "topic": "Sliding window technique",
            "example_questions": [
                "Longest substring without repeating characters.",
                "Find the minimum window substring containing all characters of pattern.",
                "Maximum sum subarray of size k.",
            ],
            "companies": ["Amazon", "Microsoft", "Adobe"],
        },
        {
            "topic": "Binary search on answer (parametric search)",
            "example_questions": [
                "Allocate minimum pages such that no student reads more than necessary.",
                "Find the minimum capacity to ship packages within D days.",
                "Koko eating bananas — find minimum speed.",
            ],
            "companies": ["Flipkart", "Amazon", "Atlassian"],
        },
        {
            "topic": "Graph BFS and DFS",
            "example_questions": [
                "Find the number of islands in a 2D grid.",
                "Clone a graph with random pointers.",
                "Minimum steps to reach target in a word-ladder problem.",
            ],
            "companies": ["Microsoft", "Amazon", "Swiggy"],
        },
        {
            "topic": "Dynamic Programming — 1D",
            "example_questions": [
                "Longest increasing subsequence.",
                "Coin change problem — minimum coins.",
                "Jump game — can you reach the last index?",
            ],
            "companies": ["Amazon", "Adobe", "Ola"],
        },
        {
            "topic": "Tree traversals and classic tree problems",
            "example_questions": [
                "Lowest common ancestor of two nodes.",
                "Diameter of a binary tree.",
                "Check if two trees are mirror images of each other.",
            ],
            "companies": ["Microsoft", "Walmart Labs", "PayPal"],
        },
        {
            "topic": "Backtracking and pruning",
            "example_questions": [
                "Generate all valid combinations of N pairs of parentheses.",
                "Solve Sudoku using backtracking.",
                "Find all subsets of a set that sum to a target.",
            ],
            "companies": ["Uber", "Atlassian", "Booking.com"],
        },
        {
            "topic": "Greedy algorithms",
            "example_questions": [
                "Activity selection / interval scheduling maximization.",
                "Minimum platforms needed at a railway station.",
                "Jump game II — minimum number of jumps.",
            ],
            "companies": ["Amazon", "Flipkart", "InMobi"],
        },
        {
            "topic": "Prefix sums and difference arrays",
            "example_questions": [
                "Count subarrays with sum equal to k.",
                "Range sum queries on a 2D matrix.",
            ],
            "companies": ["Adobe", "Intuit", "Makemytrip"],
        },
        {
            "topic": "Heap and priority queue patterns",
            "example_questions": [
                "Merge k sorted linked lists.",
                "Find the median from a data stream.",
                "Top k frequent elements.",
            ],
            "companies": ["Amazon", "Microsoft", "Nutanix"],
        },
        {
            "topic": "Union-Find and Topological Sort",
            "example_questions": [
                "Detect a cycle in a directed graph.",
                "Course schedule — can you finish all courses?",
                "Number of connected components in an undirected graph.",
            ],
            "companies": ["Atlassian", "Grab", "Uber"],
        },
    ],

    "hard": [
        # Patterns tested at Google, Meta, Rubrik, Snowflake, Stripe, Codeforces-level
        {
            "topic": "Dynamic Programming — 2D, interval, and bitmask DP",
            "example_questions": [
                "Edit distance between two strings.",
                "Burst balloons — maximize coins collected.",
                "Shortest path visiting all nodes (bitmask DP on graphs).",
            ],
            "companies": ["Google", "Meta", "Jane Street"],
        },
        {
            "topic": "Segment trees and Fenwick trees",
            "example_questions": [
                "Range minimum query with point updates.",
                "Count of smaller numbers after self.",
                "Falling squares — max height after each drop.",
            ],
            "companies": ["Rubrik", "Snowflake", "Codeforces-level"],
        },
        {
            "topic": "Advanced graph algorithms",
            "example_questions": [
                "Dijkstra's algorithm on a weighted graph with constraints.",
                "Find bridges and articulation points in a graph.",
                "Minimum cost to connect all nodes (MST variant).",
            ],
            "companies": ["Google", "Stripe", "Palantir"],
        },
        {
            "topic": "Trie data structures",
            "example_questions": [
                "Design an autocomplete system.",
                "Word search II — find all words in a board.",
                "Maximum XOR of two numbers using a binary trie.",
            ],
            "companies": ["Google", "Meta", "Rubrik"],
        },
        {
            "topic": "String algorithms — KMP, Z-function, rolling hash",
            "example_questions": [
                "Find all occurrences of a pattern in a text using KMP.",
                "Longest repeated substring using suffix arrays.",
                "Rabin-Karp for multi-pattern search.",
            ],
            "companies": ["Google", "Bloomberg", "DE Shaw"],
        },
        {
            "topic": "Monotonic stack and deque",
            "example_questions": [
                "Largest rectangle in a histogram.",
                "Sliding window maximum.",
                "Trapping rain water — both O(n) space and O(1) space solutions.",
            ],
            "companies": ["Meta", "Stripe", "Snowflake"],
        },
        {
            "topic": "Hard DP — convex hull trick and divide-and-conquer DP",
            "example_questions": [
                "Optimal strategy to minimize cost using the convex hull trick.",
                "Divide and conquer optimization for DP transitions.",
            ],
            "companies": ["Jane Street", "Two Sigma", "competitive programming"],
        },
        {
            "topic": "System design–flavored coding problems",
            "example_questions": [
                "Design an LRU cache with O(1) get and put.",
                "Implement a rate limiter using a sliding window log.",
                "Design a distributed task scheduler — key data structures.",
            ],
            "companies": ["Rubrik", "Snowflake", "Stripe", "Databricks"],
        },
        {
            "topic": "Heavy-Light Decomposition and tree queries",
            "example_questions": [
                "Path sum queries on a weighted tree with updates.",
            ],
            "companies": ["Google", "competitive programming", "ICPC-level"],
        },
        {
            "topic": "Matrix exponentiation",
            "example_questions": [
                "Compute Fibonacci(n) in O(log n) using matrix exponentiation.",
                "Count paths of length k in a graph.",
            ],
            "companies": ["Jane Street", "DE Shaw", "competitive programming"],
        },
    ],
}


# ─── Interviewer personas per difficulty ─────────────────────
# These shape the "voice" and style of the interviewer
INTERVIEWER_PERSONAS = {
    "easy": {
        "name": "Priya",
        "style": "warm, encouraging, and patient",
        "company_vibe": "a mid-size tech company doing a campus recruitment drive",
        "pacing": "Go at a comfortable pace. Give the candidate time to think.",
    },
    "medium": {
        "name": "Arjun",
        "style": "professional, direct, and thoughtful",
        "company_vibe": "a top product company like Amazon or Microsoft",
        "pacing": "Keep a steady pace. Probe deeper if the answer is shallow.",
    },
    "hard": {
        "name": "Rohan",
        "style": "sharp, concise, and intellectually demanding",
        "company_vibe": "an elite company like Google, Rubrik, or a quant fund",
        "pacing": "Move quickly. Expect precise answers. Follow up relentlessly.",
    },
}


def _get_api_keys(prefix: str) -> list[str]:
    keys = []

    exact_key = os.environ.get(prefix, "")
    if exact_key:
        keys.extend([k.strip() for k in exact_key.split(",") if k.strip()])

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
            pass

    for api_key in gemini_keys:
        url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key={api_key}"
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
    if not audio_bytes:
        return ""
        
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
    """Use gTTS for Text-to-Speech."""
    try:
        tts = gTTS(text=text, lang='en', tld='com')
        fp = io.BytesIO()
        tts.write_to_fp(fp)
        return fp.getvalue()
    except Exception as e:
        print(f"gTTS error: {e}")
        return b""


def _build_topic_context(level: str) -> str:
    """Build a concise topic list string with example questions for the prompt."""
    pool = TOPIC_POOLS.get(level, TOPIC_POOLS["easy"])
    lines = []
    for item in pool:
        examples = "; ".join(item["example_questions"][:2])
        lines.append(f"- {item['topic']} (e.g. {examples})")
    return "\n".join(lines)


def generate_first_question(level: str, duration: int) -> tuple[str, bytes]:
    """
    Generate the opening message for an AI interview session.
    Returns (question_text, audio_bytes).
    """
    persona = INTERVIEWER_PERSONAS.get(level, INTERVIEWER_PERSONAS["easy"])

    # TTS-friendly prompt: short sentences, no symbols, no lists, natural pauses via commas/periods
    prompt = f"""You are {persona['name']}, a senior software engineer at {persona['company_vibe']}.
You are {persona['style']}.
You are about to conduct a {duration}-minute technical coding interview.

Your task: deliver the opening of the interview — out loud, as if speaking to the candidate sitting across from you.

Rules you must follow:
1. Greet the candidate warmly but briefly. One sentence only.
2. Tell them the duration and rough structure: introduction, then coding questions.
3. Ask them to introduce themselves — name, background, what they have been working on recently.
4. Do NOT ask any technical question yet.
5. Total response must be under 60 words.
6. Write in short, natural spoken sentences. No bullet points, no markdown, no code, no placeholders.
7. Use a comma or period between ideas so the text-to-speech sounds natural with pauses.
8. Do not say things like "I am an AI" or reference being a bot.
9. Never use brackets or placeholder text like [Name]."""

    question_text = _call_llm(prompt)
    audio_bytes = _synthesize_speech(question_text)
    return question_text, audio_bytes


def process_response(
    level: str,
    duration: int,
    chat_history: list,
    audio_bytes: bytes,
    time_up: bool = False,
    system_action: str = "",
    code_text: str = "",
) -> tuple[str, bytes, str]:
    """
    Process the candidate's audio response or a system action,
    then generate the next interviewer turn.
    Returns (next_question_text, audio_bytes, user_transcript).
    """
    persona = INTERVIEWER_PERSONAS.get(level, INTERVIEWER_PERSONAS["easy"])

    # ── Step 1: Transcribe ───────────────────────────────────────────────
    if system_action:
        user_transcript = "[Candidate remained silent]"
    else:
        user_transcript = _transcribe_audio(audio_bytes)
        if not user_transcript.strip():
            fallback = "I did not quite catch that. Could you say that again, please?"
            return fallback, _synthesize_speech(fallback), ""

    # ── Step 2: Build conversation history string ────────────────────────
    conversation = ""
    for entry in chat_history:
        role = entry.get("role", "unknown")
        text = entry.get("text", "")
        if role == "interviewer":
            conversation += f"Interviewer: {text}\n"
        else:
            conversation += f"Candidate: {text}\n"
    conversation += f"Candidate: {user_transcript}\n"
    
    if code_text.strip():
        conversation += f"\n[Candidate's Current Code in Editor]\n{code_text}\n"

    topic_context = _build_topic_context(level)

    # ── Step 3: Choose the right prompt based on situation ───────────────

    # --- Idle nudge: candidate silent for ~1 minute ---
    if system_action == "idle_nudge":
        code_context = ""
        if code_text.strip():
            code_context = "The candidate is currently typing code but hasn't spoken. Acknowledge that you see them writing code, and ask them to talk through their thought process."
        else:
            code_context = "The candidate has gone silent for about a minute and isn't writing any code. Offer a small nudge or hint."
            
        prompt = f"""You are {persona['name']}, a senior engineer conducting a {duration}-minute coding interview.
You are {persona['style']}.

{code_context}

Say something natural and human — not robotic.
Rules:
- Under 25 words.
- Spoken, conversational tone. No markdown. No lists. No code blocks.
- Use short sentences with commas for natural TTS pacing."""

    # --- Idle skip: still silent after the nudge ---
    elif system_action == "idle_skip":
        prompt = f"""You are {persona['name']}, a senior engineer conducting a {duration}-minute coding interview.
You are {persona['style']}.

The candidate did not respond even after being prompted.

Gently acknowledge that it is okay, briefly say you will move on, and then ask a fresh question from the following topic areas:
{topic_context}

Rules:
- Under 60 words total.
- Ask only ONE new question.
- Spoken, conversational tone. Short sentences. No markdown. No lists. No code blocks.
- Make the transition feel natural, not robotic."""

    # --- Time up: wrap up the interview ---
    elif time_up:
        prompt = f"""You are {persona['name']}, a senior engineer wrapping up a {duration}-minute coding interview.
You are {persona['style']}.

Here is the full conversation:
{conversation}

Your task: deliver a natural, human closing to the interview.

Do this in order:
1. Give one or two sentences of genuine, specific feedback on the candidate's last answer — mention something concrete they said.
2. Tell them the interview is now over and thank them sincerely.
3. Optionally tell them what the next steps might be (feedback in a few days, HR will reach out, etc.).

Rules:
- Under 90 words total.
- Warm but professional tone.
- Short, spoken sentences with natural pauses.
- No markdown. No bullet points. No code blocks.
- Do not ask any new technical questions."""

    # --- Normal flow: evaluate and ask next question ---
    else:
        prompt = f"""You are {persona['name']}, a senior engineer at {persona['company_vibe']}.
You are {persona['style']}.
You are conducting a {duration}-minute technical coding interview.

Available topics and example questions you can draw from:
{topic_context}

Here is the conversation so far:
{conversation}

Your task: respond as the interviewer in the next turn.

Interview Strategy & Flow:
1. Always ask more conceptual questions and explore the candidate's thinking first.
2. Only ask them to write code if you specifically want to verify their coding skills after discussing the approach.
3. If the interview is {duration} minutes (e.g. 15 min), ask for a SMALL code snippet if you ask them to code.
4. If it's a 30 min interview, ask for a mid-sized code snippet. AVOID very tough or excessively long implementations as time is limited.

Use this decision logic:
- If the candidate just introduced themselves: ask your first conceptual question based on a topic appropriate for "{level}" difficulty. DO NOT ask them to code yet.
- If their answer is strong and correct: acknowledge it, and either ask them to code the approach (if not done yet) or move to a follow-up question.
- If their answer is partially correct: point out the gap clearly and ask a targeted follow-up.
- If their answer is wrong: correct them briefly, give a hint, and ask a simpler version.

Evaluation quality bar (use this internally to judge the answer):
- Easy level: expect correct brute force or basic optimized approach. O(n^2) is acceptable. Clean logic matters.
- Medium level: expect an optimized approach. Brute force gets partial credit. Must explain complexity.
- Hard level: expect the optimal solution with correct complexity analysis. Expect follow-up on edge cases and scalability.

Natural speech rules (critical — this will be read by a TTS engine):
- Under 80 words total.
- Short sentences. Use commas and periods to create natural spoken pauses.
- No markdown, no bullet points, no code blocks, no asterisks, no backticks.
- Do not say "Let's move on" or "Great question" or robotic filler phrases.
- Do not number questions. Ask ONE question at a time.
- Sound like a real person who is genuinely curious about how the candidate thinks."""

    next_question = _call_llm(prompt)
    next_audio = _synthesize_speech(next_question)
    return next_question, next_audio, user_transcript