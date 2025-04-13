'use client'
import React, { useEffect, useState } from 'react'
import api from '@/lib/api';
import { FaFolder, FaFolderOpen, FaFile, FaChevronRight, FaChevronDown } from 'react-icons/fa';
import Image from 'next/image';
import { Button } from './ui/button';
import { useRouter } from 'next/navigation';

interface FSNode {
    name: string;
    type: number;
    content: FSNode[] | string[];
}

interface Partition {
    partitionName: string;
    partitionType: 'P' | 'E';
    partitionSize: number;
    partitionStart: number;
    partitionFit: 'F' | 'B' | 'W';
    partitionID: string;
    fs?: FSNode;
}

interface Disk {
    diskPath: string;
    diskSize: number;
    diskSignature: string;
    diskFit: 'F' | 'B' | 'W';
    partitions: Partition[];
}

interface SelectedItem {
    type: 'disk' | 'partition' | 'file' | 'folder';
    data: any;
    path: string[];
    nodeData?: FSNode;  // Nuevo campo para almacenar los datos del nodo FSNode
}

const FileExplorer = () => {
    const [fileSystem, setFileSystem] = useState<Disk[]>([]);
    const [selectedItem, setSelectedItem] = useState<SelectedItem | null>(null);
    const [expandedItems, setExpandedItems] = useState<Record<string, boolean>>({});

    const router = useRouter(); 

    useEffect(() => {
        const fetchData = async () => {
            try {
                const command = 'getfs';
                const response = await api.post('/execute', { command });
                const jsonData = JSON.parse(response.data.output);
                setFileSystem(jsonData);
            } catch (error) {
                console.error('Error fetching data:', error);
            }
        };
        fetchData();
    }, []);

    const toggleItem = (path: string) => {
        setExpandedItems(prev => ({
            ...prev,
            [path]: !prev[path]
        }));
    };

    const handleItemClick = (type: 'disk' | 'partition' | 'file' | 'folder', data: any, path: string[], nodeData?: FSNode) => {
        setSelectedItem({ type, data, path, nodeData });
    };

    interface TreeItem {
        id: string;
        name: string;
        type: 'disk' | 'partition' | 'file' | 'folder';
        icon: React.ReactNode;
        size?: string;
        children?: React.ReactNode;
        nodeData?: FSNode;  // Nuevo campo para pasar los datos del nodo
    }

    const renderTreeItem = (item: TreeItem) => {
        const isExpanded = expandedItems[item.id];
        const isSelected = selectedItem?.path.join('/') === item.id;

        return (
            <div key={item.id}>
                <div
                    className={`flex items-center p-1 cursor-pointer ${isSelected ? 'bg-blue-100' : 'hover:bg-gray-100'}`}
                    onClick={() => {
                        handleItemClick(item.type, item, item.id.split('/'), item.nodeData);
                        if (item.type === 'disk' || item.type === 'partition' || item.type === 'folder') {
                            toggleItem(item.id);
                        }
                    }}
                >
                    {(item.type === 'disk' || item.type === 'partition' || item.type === 'folder') ? (
                        isExpanded ? (
                            <FaChevronDown className="mr-1 text-xs" />
                        ) : (
                            <FaChevronRight className="mr-1 text-xs" />
                        )
                    ) : (
                        <span className="mr-3"></span>
                    )}
                    <span className="mr-2">{item.icon}</span>
                    <span className="flex-1 truncate">{item.name}</span>
                    {item.size && <span className="text-xs text-gray-500 ml-2">{item.size}</span>}
                </div>
                {isExpanded && item.children && (
                    <div className="ml-4">
                        {item.children}
                    </div>
                )}
            </div>
        );
    };

    const renderFileSystemTree = () => {
        return (
            <div className="h-full overflow-y-auto">
                {fileSystem?.map((disk, diskIndex) => {
                    const diskId = `disk-${diskIndex}`;
                    return renderTreeItem({
                        id: diskId,
                        name: `${disk.diskPath.split('/').pop()}`,
                        type: 'disk',
                        icon: <Image src="/disk.svg" alt="Hard Drive" width={20} height={20} />,
                        size: `${Math.round(disk.diskSize / (1024 * 1024))} MB`,
                        children: expandedItems[diskId] && (
                            <>
                                {disk.partitions.map((partition, partitionIndex) => {
                                    const partitionId = `${diskId}/partition-${partitionIndex}`;
                                    return renderTreeItem({
                                        id: partitionId,
                                        name: `${partition.partitionName} (${partition.partitionType === 'P' ? 'Primaria' : partition.partitionName === 'N' ? "Disponible" : "Extendida"})`,
                                        type: 'partition',
                                        icon: <FaFolder className="text-blue-500" />,
                                        size: `${Math.round(partition.partitionSize / (1024 * 1024))} MB`,
                                        children: expandedItems[partitionId] && partition.fs && (
                                            renderFsNodes(partition.fs, partitionId)
                                        )
                                    });
                                })}
                            </>
                        )
                    });
                })}
            </div>
        );
    };

    const renderFsNodes = (node: FSNode, parentPath: string): React.ReactNode => {
        const nodePath = `${parentPath}/${node.name}`;
        const isDirectory = node.type === 0;

        return renderTreeItem({
            id: nodePath,
            name: node.name,
            type: isDirectory ? 'folder' : 'file',
            icon: isDirectory ? 
                (expandedItems[nodePath] ? 
                    <FaFolderOpen className="text-yellow-500" /> : 
                    <FaFolder className="text-yellow-500" />) : 
                <FaFile className="text-blue-400" />,
            children: expandedItems[nodePath] && isDirectory && Array.isArray(node.content) && (
                <div>
                    {(node.content as FSNode[]).map((child, index) => 
                        renderFsNodes(child, nodePath)
                    )}
                </div>
            ),
            nodeData: node  // Pasamos el nodo completo con su contenido
        });
    };

    const renderContent = () => {
        if (!selectedItem) {
            return (
                <div className="flex items-center justify-center h-full text-gray-500">
                    Seleccione un elemento para ver su contenido
                </div>
            );
        }

        switch (selectedItem.type) {
            case 'disk':
                const disk = fileSystem[parseInt(selectedItem.path[0].split('-')[1])];
                return (
                    <div className="p-4">
                        <h3 className="text-lg font-bold mb-4">Información del Disco</h3>
                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <p className="font-medium">Ruta:</p>
                                <p>{disk.diskPath}</p>
                            </div>
                            <div>
                                <p className="font-medium">Tamaño:</p>
                                <p>{Math.round(disk.diskSize / (1024 * 1024))} MB</p>
                            </div>
                            <div>
                                <p className="font-medium">Firma:</p>
                                <p>{disk.diskSignature}</p>
                            </div>
                            <div>
                                <p className="font-medium">Ajuste:</p>
                                <p>
                                    {disk.diskFit === 'F' ? 'First Fit' : 
                                     disk.diskFit === 'B' ? 'Best Fit' : 'Worst Fit'}
                                </p>
                            </div>
                        </div>
                        <h4 className="text-lg font-bold mt-6 mb-2">Particiones</h4>
                        <div className="space-y-2">
                            {disk.partitions.map((partition, index) => (
                                <div key={index} className="border p-2 rounded">
                                    <p className="font-medium">{partition.partitionName}</p>
                                    <p>Tipo: {partition.partitionType === 'P' ? 'Primaria' : 'Extendida'}</p>
                                    <p>Tamaño: {Math.round(partition.partitionSize / (1024 * 1024))} MB</p>
                                </div>
                            ))}
                        </div>
                    </div>
                );

            case 'partition':
                const diskIndex = parseInt(selectedItem.path[0].split('-')[1]);
                const partitionIndex = parseInt(selectedItem.path[1].split('-')[1]);
                const partition = fileSystem[diskIndex].partitions[partitionIndex];
                return (
                    <div className="p-4">
                        <h3 className="text-lg font-bold mb-4">Información de la Partición</h3>
                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <p className="font-medium">Nombre:</p>
                                <p>{partition.partitionName}</p>
                            </div>
                            <div>
                                <p className="font-medium">Tipo:</p>
                                <p>{partition.partitionType === 'P' ? 'Primaria' : 'Extendida'}</p>
                            </div>
                            <div>
                                <p className="font-medium">Tamaño:</p>
                                <p>{Math.round(partition.partitionSize / (1024 * 1024))} MB</p>
                            </div>
                            <div>
                                <p className="font-medium">Ajuste:</p>
                                <p>
                                    {partition.partitionFit === 'F' ? 'First Fit' : 
                                     partition.partitionFit === 'B' ? 'Best Fit' : 'Worst Fit'}
                                </p>
                            </div>
                            <div>
                                <p className="font-medium">Inicio:</p>
                                <p>{partition.partitionStart}</p>
                            </div>
                            <div>
                                <p className="font-medium">ID:</p>
                                <p>{partition.partitionID}</p>
                            </div>
                        </div>
                    </div>
                );

            case 'folder':
                return (
                    <div className="p-4">
                        <h3 className="text-lg font-bold mb-4">Carpeta {selectedItem.data.name}</h3>
                        {selectedItem.data.children && (
                            <p>Esta carpeta contiene {React.Children.count(selectedItem.data.children)} elementos</p>
                        )}
                    </div>
                );

            case 'file':
                // Accedemos al contenido a través de nodeData en lugar de data
                const fileContent = selectedItem.nodeData?.content;
                return (
                    <div className="p-4">
                        <h3 className="text-lg font-bold mb-4">{selectedItem.nodeData?.name}</h3>
                        <div className="bg-gray-100 p-4 rounded">
                            {Array.isArray(fileContent) && fileContent.length > 0 ? (
                                <pre className="whitespace-pre-wrap">
                                    {fileContent.join('\n')}
                                </pre>
                            ) : (
                                <div className="text-gray-500">Archivo vacío</div>
                            )}
                        </div>
                    </div>
                );

            default:
                return null;
        }
    };

    return (
        <div className="flex flex-col h-full">
            <div className="flex items-center justify-between p-4 border-b bg-slate-200">
                <h2 className="text-xl font-bold">Explorador de Archivos</h2>
                <Button onClick={() => router.push("/")} className="bg-blue-500 hover:bg-blue-600 text-white">Volver al inicio</Button>
            </div>
            <div className="flex flex-1 overflow-hidden">
                {/* Panel izquierdo - Árbol de navegación */}
                <div className="w-1/3 border-r overflow-y-auto bg-gray-50">
                    {renderFileSystemTree()}
                </div>
                
                {/* Panel derecho - Contenido */}
                <div className="flex-1 overflow-y-auto bg-white">
                    {renderContent()}
                </div>
            </div>
        </div>
    );
};

export default FileExplorer;