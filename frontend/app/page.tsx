"use client";

import { useState } from "react";
import Terminal from "../components/Terminal";
import api from "@/lib/api";
import FileUpload from "@/components/FileUpload";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { toast } from "@/hooks/use-toast";

export default function FileSystemSimulator() {
  const [inputContent, setInputContent] = useState<string[]>([""]);
  const [outputContent, setOutputContent] = useState<string[]>([]);

  // Función para manejar cambios en el contenido
  const handleContentChange = (index: number, newValue: string) => {
    const newContent = [...inputContent];
    newContent[index] = newValue;
    setInputContent(newContent);
  };

  // Función para manejar la ejecución de comandos
  const handleExecute = async (command: string) => {
    if (command === "clear" || command === "cls") {
      setOutputContent([]);
      setInputContent([""]);
      return;
    }
    if (command.trim() !== "") {
      try {
        // Enviar el comando al servidor local usando la instancia de axios
        const response = await api.post("/execute", {
          command: command,
        });

        const output = response.data.output;

        // mostrar cada linea de salida
        for (const line of output.split("\n")) {
          setOutputContent((prev) => [...prev, line]);
        }

        // Añadir el comando ejecutado al contenido de salida
        setOutputContent((prev) => [`> ${command}`, ...prev]);
        // Añadir el resultado al contenido de salida

      } catch (error) {
        // Manejar errores de la petición
        setOutputContent([...outputContent, `> ${command}`, "Error: Could not execute command"]);
        console.error("Error executing command:", error);
      }

      // Añadir una nueva línea vacía para seguir escribiendo
      setInputContent([...inputContent, ""]);
    }
  };

  const handleUpload = (content: string) => {
    console.log("Contenido del archivo subido:", content);
    for (const line of content.split("\n")) {
      setOutputContent((prev) => [...prev, line]);
    }
    setInputContent((prev) => [...prev, ""]);

    handleExecute(content);
  };

  const handleLogout = async () => {
    const command = "logout";
    const response = await api.post("/execute", {
      command: command
    });
    console.log('Response:', response);
    toast({
      title: 'Logout',
      description: response.data.output,
      variant: 'default'
    })
  };

  return (
    <div className="min-h-screen p-8"
      style={{ fontFamily: "'Fira Code', monospace", backgroundColor: "#202020" }}
    >
      <h1 className="text-4xl font-bold mb-8 text-center text-white">File System EXT2</h1>

      <Terminal
        content={inputContent}
        contentOutput={outputContent}
        editable={true}
        onContentChange={handleContentChange}
        onExecute={handleExecute}
      />

      <div className="flex justify-center mb-4">
        <label
          htmlFor="file-upload"
          className="px-4 py-2 rounded-lg bg-blue-500 text-white hover:bg-blue-600 
                       transition-colors duration-200 cursor-pointer shadow-sm
                       flex items-center gap-2"
        >
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
            />
          </svg>
          Cargar Archivo
        </label>
        <Button variant="outline" className="ml-4 bg-green-500 text-white hover:bg-green-600" onClick={() => handleExecute(inputContent[inputContent.length - 1])}>
          Ejecutar
        </Button>
        <Link href="/file-explorer">
          <Button variant="outline" className="ml-4 bg-blue-500 text-white hover:bg-blue-600">File Explorer</Button>
        </Link>
        <Link href="/login">
          <Button variant="outline" className="ml-4">Login</Button>
        </Link>
        <Button variant="outline" className="ml-4 bg-red-500 text-white hover:bg-red-600" onClick={handleLogout}>Logout</Button>
      </div>

      <div className="flex justify-center space-x-4">
        <FileUpload onFileContent={handleUpload} />
      </div>
    </div>
  );
}