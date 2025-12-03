
import { GoogleGenAI, Type, Schema } from "@google/genai";
import { GeneratedResponse, CodeFile, ChatMessage } from "../types";

const ai = new GoogleGenAI({ apiKey: process.env.API_KEY });

const responseSchema: Schema = {
  type: Type.OBJECT,
  properties: {
    message: {
      type: Type.STRING,
      description: "Conversational response to the user. Use Markdown. Explain technical decisions and fixes clearly. (Write in Persian/Farsi)"
    },
    files: {
      type: Type.ARRAY,
      description: "List of files that need to be created or UPDATED. If a file is unchanged, do not include it here.",
      items: {
        type: Type.OBJECT,
        properties: {
          filename: { type: Type.STRING, description: "Name of the file including extension (e.g., src/main.py)" },
          language: { type: Type.STRING, description: "Programming language for syntax highlighting" },
          content: { type: Type.STRING, description: "The FULL, VALID content of the file." }
        },
        required: ["filename", "language", "content"]
      }
    },
    databaseInstructions: {
      type: Type.STRING,
      description: "Technical instructions for database setup (SQL, migrations, connection strings) if needed, otherwise null."
    }
  },
  required: ["message", "files"]
};

export const generateCode = async (
  currentFiles: CodeFile[],
  userPrompt: string,
  history: ChatMessage[]
): Promise<GeneratedResponse> => {
  const modelId = "gemini-3-pro-preview";

  // Filter massive history to prevent token overflow, keep last 10 messages
  const recentHistory = history.slice(-10).map(msg => `${msg.role.toUpperCase()}: ${msg.content}`).join('\n');

  const fileContext = currentFiles.length > 0 
    ? `
      === CURRENT PROJECT STATE ===
      The following files currently exist in the project. 
      Use this context to FIX BUGS, REFACTOR, or ADD FEATURES without breaking existing functionality.
      
      ${JSON.stringify(currentFiles.map(f => ({ filename: f.filename, content: f.content })))}
      =============================
      `
    : `
      === NEW PROJECT ===
      No files exist yet. Generate a complete, working project structure from scratch.
      `;

  const systemInstruction = `
    You are CodeMaster AI, an elite Senior Software Architect and Polyglot Developer.
    
    YOUR MISSION:
    Turn user ideas into production-grade, bug-free, and secure code. 
    You are not just a code generator; you are an intelligent maintainer.
    
    CORE RULES:
    1. **Expert Quality**: Write clean, DRY, modular code following SOLID principles. Use modern syntax (e.g., ES6+ for JS, Type Hints for Python).
    2. **Dependency Management**: IF external libraries are needed, YOU MUST generate the dependency file (package.json, requirements.txt, go.mod, pom.xml, etc.).
    3. **Debugging & Fixing**: If the user reports an error, analyze the 'CURRENT PROJECT STATE'. Find the root cause and return the *corrected* version of the specific file(s).
    4. **Smart Updates**: Only return files that have changed or are new. Do not return unchanged files.
    5. **Language**:
       - Code: English (Standard naming conventions).
       - Chat/Explanations: **Persian (Farsi)**. Tone: Professional, authoritative, yet helpful.
    
    RESPONSE FORMAT:
    - **message**: Explain WHAT you did and WHY. If fixing a bug, explain the fix. Use Markdown (bold, code blocks) for readability.
    - **files**: The array of file objects.
    - **databaseInstructions**: If a DB is involved, explain how to set it up (e.g., "Run docker-compose up" or "Execute schema.sql").
  `;

  const finalPrompt = `
    ${fileContext}

    CHAT HISTORY:
    ${recentHistory}

    USER REQUEST: "${userPrompt}"
    
    (Remember: Return JSON matching the schema. If fixing a bug, update the code directly.)
  `;

  try {
    const result = await ai.models.generateContent({
      model: modelId,
      contents: finalPrompt,
      config: {
        systemInstruction: systemInstruction,
        responseMimeType: "application/json",
        responseSchema: responseSchema,
        temperature: 0.2, // Low temperature for precise code generation
      }
    });

    const text = result.text;
    if (!text) {
      throw new Error("No response generated from AI.");
    }

    return JSON.parse(text) as GeneratedResponse;
  } catch (error) {
    console.error("Gemini Generation Error:", error);
    throw error;
  }
};
