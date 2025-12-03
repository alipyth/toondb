
export interface CodeFile {
  filename: string;
  language: string;
  content: string;
}

export interface ChatMessage {
  role: 'user' | 'model';
  content: string;
  timestamp: number;
}

export interface GeneratedResponse {
  message: string; // The chat response from AI
  files: CodeFile[]; // The files that were created or updated
  databaseInstructions: string | null;
}

export enum AppState {
  IDLE = 'IDLE',
  GENERATING = 'GENERATING',
  ERROR = 'ERROR'
}
