import Editor from '@monaco-editor/react';
import './CodeEditor.css';

const DEFAULT_TEMPLATE = `#include <bits/stdc++.h>
using namespace std;

int main() {
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    
    // Your code here
    
    return 0;
}
`;

export default function CodeEditor({ value, onChange, language = 'cpp', options = {} }) {
  const handleEditorChange = (val) => {
    if (onChange) onChange(val);
  };

  return (
    <div className="code-editor-wrapper" id="code-editor">
      <Editor
        height="100%"
        language={language}
        theme="vs-dark"
        value={value || DEFAULT_TEMPLATE}
        onChange={handleEditorChange}
        options={{
          fontSize: 13,
          fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
          minimap: { enabled: false },
          scrollBeyondLastLine: false,
          padding: { top: 12 },
          lineNumbers: 'on',
          roundedSelection: true,
          renderLineHighlight: 'line',
          selectOnLineNumbers: true,
          wordWrap: 'on',
          automaticLayout: true,
          tabSize: 4,
          suggestOnTriggerCharacters: true,
          bracketPairColorization: { enabled: true },
          cursorBlinking: 'smooth',
          cursorSmoothCaretAnimation: 'on',
          smoothScrolling: true,
          ...options
        }}
      />
    </div>
  );
}
