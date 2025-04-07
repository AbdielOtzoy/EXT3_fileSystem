import { useEffect, useRef } from "react";

interface TerminalProps {
  content: string[];
  contentOutput?: string[];
  editable?: boolean;
  onContentChange?: (index: number, newValue: string) => void;
  onExecute?: (command: string) => void;
}

const Terminal: React.FC<TerminalProps> = ({
  content,
  contentOutput,
  editable = false,
  onContentChange,
  onExecute,
}) => {
  const terminalRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [content, contentOutput]);

  const handleChange = (index: number) => (event: React.ChangeEvent<HTMLInputElement>) => {
    if (onContentChange) {
      onContentChange(index, event.target.value);
    }
  };

  const handleKeyDown = (index: number) => (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter" && onExecute) {
      onExecute(content[index]);
    }
  };

  return (
    <div className="flex flex-col mb-5">
      <div className="bg-gray-800 text-white p-2 rounded-t-lg flex justify-between items-center">
        <div className="flex space-x-2 p-3">
          <div className="w-3 h-3 bg-red-500 rounded-full"></div>
          <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
          <div className="w-3 h-3 bg-green-500 rounded-full"></div>
        </div>
      </div>
      <div
        ref={terminalRef}
        className="flex-grow bg-black text-white p-4 rounded-b-lg overflow-auto font-mono h-[400px]"
      >
        {contentOutput &&
          contentOutput.map((line, index) => (
            <div key={index} className="flex">
              <span className="text-gray-500 mr-4 select-none">{index + 1}</span>
              <span>{line}</span>
            </div>
          ))}
        {editable && (
          <div className="flex">
            <span className="text-gray-500 mr-4 select-none">{contentOutput?.length || 0 + 1}</span>
            <input
              type="text"
              value={content[content.length - 1] || ""}
              onChange={handleChange(content.length - 1)}
              onKeyDown={handleKeyDown(content.length - 1)}
              className="bg-transparent outline-none flex-grow"
              autoFocus
            />
          </div>
        )}
      </div>
    </div>
  );
};

export default Terminal;