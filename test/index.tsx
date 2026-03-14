import React, { useState, useEffect, useRef } from 'react';
import { Leva, useControls, button, folder, useCreateStore } from 'leva';
import { App, Dropdown } from 'antd';
import type { DesktopPlugin } from '../../../../../types/Plugin';
import { SourceCodeEditor } from './SourceCodeEditor';
import { useDraggable } from '../../../../../hooks/useDraggable';
import { LevaOverrides } from './DevWorkbench.styles';
import { compilePluginSource, extractPluginClassName, getFirstErrorMessage } from '../../Plugins/DynamicLoader/pluginSourceUtils';
import {
  WorkbenchContainer,
  Title,
  CloseBtn,
  FileUploadArea,
  FileInput,
  PluginStats,
  StatusDot,
  ActionArea
} from './DevWorkbench.styles';
import { VisualBuilder } from './FunPanel';
import { autoRunTsPlugin, autoRunPlugin, autoRunClassicalLampTsPlugin } from './testRunPlugin';
import { FileExplorer } from './FileExplorer';
import {
    PlusCircleOutlined,
    StopOutlined,
    PlayCircleOutlined,
    UploadOutlined,
    DeleteOutlined,
    WarningOutlined,
    SaveOutlined
} from '@ant-design/icons';
import { PluginFileNode, DevPluginData } from '../../../../../types/devWorkbench';
import { fileExplorerApi } from '../../../../../api/fileExplorerApi';

interface DevWorkbenchProps {
    pluginManager: any | null; // We'll pass the PluginManager instance
    onClose?: () => void;
}

export const DevWorkbench: React.FC<DevWorkbenchProps> = ({ pluginManager, onClose }) => {
    const containerRef = useRef<HTMLDivElement>(null);
    const handleRef = useRef<HTMLHeadingElement>(null);
    useDraggable(containerRef, handleRef);

    const { message: messageApi } = App.useApp();
    const fileInputRef = useRef<HTMLInputElement>(null);
    const nestedFileInputRef = useRef<HTMLInputElement>(null);
    const [nestedFileUploadPath, setNestedFileUploadPath] = useState<string | null>(null);

    const handleNestedFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (!file || !nestedFileUploadPath || !fileTree) return;
        const reader = new FileReader();
        reader.onload = async (e) => {
            const result = e.target?.result as string;
            if (result === undefined) return;
            const newNode: PluginFileNode = { name: file.name, type: 'file', content: result };
            const updatedTree = addNodeToTree(fileTree, nestedFileUploadPath, fileTree.name, newNode);
            updateSelectedPluginState(p => ({ ...p, fileTree: updatedTree }));
            setNestedFileUploadPath(null);
        };
        reader.readAsText(file);
        if (nestedFileInputRef.current) nestedFileInputRef.current.value = '';
    };
    const [isHovering, setIsHovering] = useState(false);
    
    // Plugin List UI state
    const [devPlugins, setDevPlugins] = useState<DevPluginData[]>([]);
    const [selectedPluginId, setSelectedPluginId] = useState<string | null>(null);
    const [hoveredPlugin, setHoveredPlugin] = useState<string | null>(null);

    // Derived states
    const selectedPlugin = devPlugins.find(p => p.id === selectedPluginId) || null;
    const loadedPluginId = selectedPlugin?.id || null;
    const activePlugin = selectedPlugin?.instance || null;
    const pluginSource = selectedPlugin?.source || '';
    const currentFilename = selectedPlugin?.currentFilename || '';
    const fileTree = selectedPlugin?.fileTree || null;

    const [showEditor, setShowEditor] = useState<boolean>(false);
    const [showVisualBuilder, setShowVisualBuilder] = useState<boolean>(false);
    
    // File Manager State
    const [hoveredNode, setHoveredNode] = useState<string | null>(null);
    const [selectedNodePath, setSelectedNodePath] = useState<string | null>(null);
    const [editingNodePath, setEditingNodePath] = useState<string | null>(null);
    const [editingNodeName, setEditingNodeName] = useState<string>('');
    const [addingNodePath, setAddingNodePath] = useState<string | null>(null);
    const [addingNodeType, setAddingNodeType] = useState<'file' | 'folder' | null>(null);
    const [addingNodeName, setAddingNodeName] = useState<string>('');

    /**
     * 将文件树拍平成源码文件表，方便编译器和热更新复用。
     */
    const treeToSourceFiles = (node: PluginFileNode | null, currentPath = ''): Record<string, string> => {
        if (!node) return {};
        const nextPath = currentPath ? `${currentPath}/${node.name}` : node.name;
        if (node.type === 'file') {
            return node.content !== undefined ? { [nextPath]: node.content } : {};
        }

        return (node.children || []).reduce<Record<string, string>>((acc, child) => {
            return { ...acc, ...treeToSourceFiles(child, nextPath) };
        }, {});
    };

    /**
     * 去掉根目录名，只保留插件内部相对路径。
     */
    const normalizePluginSourceFiles = (sourceFiles: Record<string, string>, rootName: string) => {
        const normalized: Record<string, string> = {};
        for (const [filePath, content] of Object.entries(sourceFiles)) {
            const normalizedPath = filePath.startsWith(`${rootName}/`)
                ? filePath.slice(rootName.length + 1)
                : filePath;
            normalized[normalizedPath] = content;
        }
        return normalized;
    };

    /**
     * 在插件文件集中推断入口文件，优先 index.ts / index.js。
     */
    const resolveEntryFilename = (sourceFiles: Record<string, string>, fallback = 'index.js') => {
        const candidates = ['index.ts', 'index.js', 'main.ts', 'main.js', fallback];
        return candidates.find(candidate => sourceFiles[candidate] !== undefined) || Object.keys(sourceFiles)[0] || fallback;
    };

    const handleSavePlugin = async (pluginId: string) => {
        const plugin = devPlugins.find(p => p.id === pluginId);
        if (!plugin || !plugin.fileTree) return;

        const formData = new FormData();
        const filesMap = treeToSourceFiles(plugin.fileTree);
        
        Object.entries(filesMap).forEach(([path, content]) => {
            formData.append('files', new Blob([content], { type: 'text/plain' }), path);
        });

        formData.append('fileTree', JSON.stringify(plugin.fileTree));

        try {
            messageApi.loading({ content: '保存中...', key: 'savePlugin' });
            const res = await fileExplorerApi.savePlugin(pluginId, formData);
            if (res.code === 200) {
                messageApi.success({ content: '保存成功！', key: 'savePlugin' });
                const newId = (typeof res.data === 'object' && res.data !== null && 'id' in res.data) ? res.data.id : res.data;
                if (typeof newId === 'string' && newId !== pluginId) {
                    setDevPlugins(prev => prev.map(p => p.id === pluginId ? { ...p, id: newId } : p));
                    if (selectedPluginId === pluginId) {
                        setSelectedPluginId(newId);
                    }
                }
            } else {
                messageApi.error({ content: res.msg || '保存失败', key: 'savePlugin' });
            }
        } catch(e: any) {
            messageApi.error({ content: e.message || '保存失败', key: 'savePlugin' });
        }
    };

    /**
     * 更新文件树中的某个文件内容，用于编辑器保存后的热更新。
     */
    const updateNodeContentInTree = (tree: PluginFileNode, targetPath: string, currentPath: string, newContent: string): PluginFileNode => {
        if (currentPath === targetPath && tree.type === 'file') {
            return { ...tree, content: newContent };
        }
        if (tree.children) {
            return {
                ...tree,
                children: tree.children.map(child => updateNodeContentInTree(child, targetPath, `${currentPath}/${child.name}`, newContent))
            };
        }
        return tree;
    };

    /**
     * 根据 sourceFiles 重建目录树，支持上传整个插件文件夹。
     */
    const buildFileTreeFromSourceFiles = (rootName: string, sourceFiles: Record<string, string>): PluginFileNode => {
        const root: PluginFileNode = {
            name: rootName,
            type: 'folder',
            isOpen: true,
            children: [],
        };

        const ensureFolder = (children: PluginFileNode[], folderName: string) => {
            let folderNode = children.find(child => child.type === 'folder' && child.name === folderName);
            if (!folderNode) {
                folderNode = { name: folderName, type: 'folder', isOpen: true, children: [] };
                children.push(folderNode);
            }
            return folderNode;
        };

        Object.entries(sourceFiles).forEach(([relativePath, content]) => {
            const parts = relativePath.split('/').filter(Boolean);
            if (parts.length === 0) return;

            let cursor = root;
            for (let index = 0; index < parts.length - 1; index += 1) {
                cursor.children = cursor.children || [];
                cursor = ensureFolder(cursor.children, parts[index]);
            }

            cursor.children = cursor.children || [];
            cursor.children.push({
                name: parts[parts.length - 1],
                type: 'file',
                content,
            });
        });

        return root;
    };

    // helper to update state of active plugin
    const updateSelectedPluginState = (updater: (prev: DevPluginData) => DevPluginData) => {
        if (!selectedPluginId) return;
        setDevPlugins(prev => prev.map(p => p.id === selectedPluginId ? updater(p) : p));
    };

    // Replace F2 to rename
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'F2' && selectedNodePath && selectedNodePath !== editingNodePath) {
                e.preventDefault();
                setEditingNodePath(selectedNodePath);
                const name = selectedNodePath.split('/').pop() || '';
                setEditingNodeName(name);
            }
        };
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [selectedNodePath, editingNodePath]);

    const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const files = event.target.files;
        if (!files || files.length === 0) return;

        let manifestFile: File | undefined;
        const fileEntries: Array<{ file: File; relativePath: string }> = [];

        for (let i = 0; i < files.length; i += 1) {
            const file = files[i];
            const relativePath = ((file as any).webkitRelativePath || file.name || '').replace(/^\/+/, '');
            const normalizedPath = relativePath || file.name;
            fileEntries.push({ file, relativePath: normalizedPath });
            if (normalizedPath.toLowerCase().endsWith('manifest.json') && !manifestFile) {
                manifestFile = file;
            }
        }

        const firstSegments = fileEntries.map(item => item.relativePath.split('/')[0]).filter(Boolean);
        const sharedRoot = firstSegments.length > 0 && firstSegments.every(segment => segment === firstSegments[0])
            ? firstSegments[0]
            : null;
        const normalizedEntries = fileEntries.map(item => {
            if (sharedRoot && item.relativePath.startsWith(`${sharedRoot}/`)) {
                return { ...item, relativePath: item.relativePath.slice(sharedRoot.length + 1) };
            }
            return item;
        });

        const entryCandidates = normalizedEntries.filter(({ relativePath }) => /(^|\/)(index|main)\.(js|ts)$/i.test(relativePath));
        const fallbackCandidates = normalizedEntries.filter(({ relativePath }) => /\.(js|ts)$/i.test(relativePath));
        const entryFile = (entryCandidates[0] || fallbackCandidates[0])?.file;
        const entryPath = (entryCandidates[0] || fallbackCandidates[0])?.relativePath;

        if (!entryFile || !entryPath) {
            messageApi.error('未找到插件入口文件 (.js / .ts)');
            return;
        }

        const readFileText = (f: File): Promise<string> => {
            return new Promise((resolve) => {
                const reader = new FileReader();
                reader.onload = (e) => resolve((e.target?.result as string) || '');
                reader.readAsText(f);
            });
        };

        try {
            const entryCode = await readFileText(entryFile);
            const sourceFilesMap: Record<string, string> = {};

            for (const { file, relativePath } of normalizedEntries) {
                const content = await readFileText(file);
                sourceFilesMap[relativePath] = content;
            }

            const compileResult = compilePluginSource(entryCode, entryPath, sourceFilesMap);
            const firstError = getFirstErrorMessage(compileResult.diagnostics);
            if (firstError) {
                throw new Error(firstError);
            }

            let pluginName = entryPath.split('/').pop()?.replace(/\.(js|ts)$/i, '') || 'plugin';
            const extractedName = extractPluginClassName(entryCode, entryPath);
            if (extractedName) {
                pluginName = extractedName;
            }
            
            if (manifestFile) {
                const manifestText = await readFileText(manifestFile);
                try {
                    const manifest = JSON.parse(manifestText);
                    if (manifest.name) pluginName = manifest.name;
                } catch(e) { }
            }

            const pluginId = `DevPlugin-${pluginName}@${Date.now()}`;
            const loadedInstance = await pluginManager.loadDynamicPlugin(entryCode, pluginId, { id: pluginId }, entryPath, sourceFilesMap);
            
            if (loadedInstance) {
                const newPlugin: DevPluginData = {
                    id: pluginId,
                    name: pluginName,
                    source: entryCode,
                    currentFilename: entryPath,
                    instance: loadedInstance,
                    fileTree: buildFileTreeFromSourceFiles(pluginName, sourceFilesMap)
                };
                setDevPlugins(prev => {
                    const existing = prev.filter(p => p.id !== pluginId);
                    return [...existing, newPlugin];
                });
                setSelectedPluginId(pluginId);
            }
        } catch (err: any) {
            messageApi.error(`Plugin load failed: ${err.message}`);
            console.error(err);
        }
        
        if (fileInputRef.current) fileInputRef.current.value = '';
    };

    useEffect(() => {
        if (pluginManager && devPlugins.length === 0) {
            fileExplorerApi.getPluginList().then(async (res) => {
                if (res.code === 200 && res.data && res.data.length > 0) {
                    const list = res.data;
                    const loadedPlugins: DevPluginData[] = [];
                    for (const info of list) {
                        try {
                            const treeRes = await fileExplorerApi.getPluginTree(info.id);
                            if (treeRes.code === 200 && treeRes.data) {
                                const tree = treeRes.data;
                                const sourceFilesMap = treeToSourceFiles(tree);
                                const normalizedEntries = Object.keys(sourceFilesMap).map(p => ({ relativePath: p, content: sourceFilesMap[p] }));
                                const entryCandidates = normalizedEntries.filter(i => /(^|\/)(index|main)\.(js|ts)$/i.test(i.relativePath));
                                const fallbackCandidates = normalizedEntries.filter(i => /\.(js|ts)$/i.test(i.relativePath));
                                const entryItem = entryCandidates[0] || fallbackCandidates[0];

                                if (entryItem) {
                                    let entryCode = entryItem.content;
                                    let entryPath = entryItem.relativePath.replace(`${tree.name}/`, '');
                                    const sourceFilesOnly = normalizePluginSourceFiles(sourceFilesMap, tree.name);
                                    const loadedInstance = await pluginManager.loadDynamicPlugin(entryCode, info.id, { id: info.id }, entryPath, sourceFilesOnly);
                                    loadedPlugins.push({
                                        id: info.id,
                                        name: info.name,
                                        source: entryCode,
                                        currentFilename: entryPath,
                                        instance: loadedInstance || null,
                                        fileTree: tree
                                    });
                                }
                            }
                        } catch (err) {
                            console.error(`Failed to load plugin tree for ${info.id}`, err);
                        }
                    }
                    if (loadedPlugins.length > 0) {
                        setDevPlugins(loadedPlugins);
                        setSelectedPluginId(loadedPlugins[0].id);
                    }
                } else {
                    // Fallback to auto run if no plugins returned
                    const appendPlugin = (pluginData: DevPluginData | null) => {
                        if (!pluginData) return;
                        setDevPlugins(prev => {
                            const existing = prev.filter(p => p.id !== pluginData.id);
                            return [...existing, pluginData];
                        });
                        setSelectedPluginId(current => current || pluginData.id);
                    };

                    autoRunTsPlugin(pluginManager).then(appendPlugin).catch(console.error);
                    autoRunClassicalLampTsPlugin(pluginManager).then(appendPlugin).catch(console.error);
                }
            }).catch(err => {
                console.error("Failed to fetch plugin list from server", err);
            });
        }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pluginManager]);

    // Handle F2 shortcut for renaming and global click outside for force blur
    useEffect(() => {
        const handleGlobalPointerDown = (e: PointerEvent) => {
            const ae = document.activeElement as HTMLElement | null;
            if (ae && ae.tagName === 'INPUT' && ae.classList.contains('workbench-edit-input')) {
                if (e.target !== ae) {
                    ae.blur();
                }
            }
        };
        const handleGlobalKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'F2' && selectedNodePath && !editingNodePath) {
                const parts = selectedNodePath.split('/');
                const name = parts[parts.length - 1];
                setEditingNodePath(selectedNodePath);
                setEditingNodeName(name);
            }
        };
        window.addEventListener('keydown', handleGlobalKeyDown);
        document.addEventListener('pointerdown', handleGlobalPointerDown, true);
        return () => {
            window.removeEventListener('keydown', handleGlobalKeyDown);
            document.removeEventListener('pointerdown', handleGlobalPointerDown, true);
        };
    }, [selectedNodePath, editingNodePath]);

    const toggleFolderOpen = (tree: PluginFileNode, targetPath: string, currentPath: string): PluginFileNode => {
        if (currentPath === targetPath && tree.type === 'folder') {
            return { ...tree, isOpen: !tree.isOpen };
        }
        if (tree.children) {
            return {
                ...tree,
                children: tree.children.map(child => toggleFolderOpen(child, targetPath, `${currentPath}/${child.name}`))
            };
        }
        return tree;
    };

    const renameNodeInTree = (tree: PluginFileNode, targetPath: string, currentPath: string, newName: string): PluginFileNode => {
        if (currentPath === targetPath) {
            return { ...tree, name: newName };
        }
        if (tree.children) {
            return {
                ...tree,
                children: tree.children.map(child => renameNodeInTree(child, targetPath, `${currentPath}/${child.name}`, newName))
            };
        }
        return tree;
    };

    const handleRenameSubmit = () => {
        if (!editingNodePath || !editingNodeName.trim() || !fileTree) {
            setEditingNodePath(null);
            return;
        }
        const updatedTree = renameNodeInTree(fileTree, editingNodePath, fileTree.name, editingNodeName.trim());
        const isRoot = editingNodePath === fileTree.name;
        
        updateSelectedPluginState(p => ({
            ...p,
            fileTree: updatedTree,
            ...(isRoot ? { name: editingNodeName.trim() } : {})
        }));

        if (selectedPluginId && !selectedPluginId.startsWith('DevPlugin-')) {
            fileExplorerApi.rename(selectedPluginId, editingNodePath, editingNodeName.trim());
        }

        setEditingNodePath(null);
    };

    const addNodeToTree = (tree: PluginFileNode, targetPath: string, currentPath: string, newNode: PluginFileNode): PluginFileNode => {
        if (targetPath === currentPath) {
            return {
                ...tree,
                isOpen: true,
                children: [newNode, ...(tree.children || [])]
            };
        }
        if (tree.children) {
            return {
                ...tree,
                children: tree.children.map(child => addNodeToTree(child, targetPath, `${currentPath}/${child.name}`, newNode))
            };
        }
        return tree;
    };

    const deleteNodeFromTree = (tree: PluginFileNode, targetPath: string, currentPath: string): PluginFileNode | null => {
        if (targetPath === currentPath) {
            return null; // Node to delete
        }
        if (tree.children) {
            return {
                ...tree,
                children: tree.children
                    .map(child => deleteNodeFromTree(child, targetPath, `${currentPath}/${child.name}`))
                    .filter((child): child is PluginFileNode => child !== null)
            };
        }
        return tree;
    };

    const handleDeleteNode = (pathToDelete: string) => {
        if (window.confirm("确定要删除此文件吗？")) {
            if (fileTree) {
                const updatedTree = deleteNodeFromTree(fileTree, pathToDelete, fileTree.name);
                if (updatedTree) {
                    updateSelectedPluginState(p => ({ ...p, fileTree: updatedTree }));
                    if (selectedPluginId && !selectedPluginId.startsWith('DevPlugin-')) {
                        fileExplorerApi.delete(selectedPluginId, pathToDelete);
                    }
                }
            }
        }
    };

    const handleAddSubmit = () => {
        if (!addingNodePath || !fileTree) {
            setAddingNodePath(null);
            setAddingNodeType(null);
            return;
        }
        const trimmed = addingNodeName.trim();
        if (trimmed && addingNodeType) {
            const newNode: PluginFileNode = { name: trimmed, type: addingNodeType };
            if (addingNodeType === 'folder') {
                newNode.children = [];
                newNode.isOpen = true;
                if (selectedPluginId && !selectedPluginId.startsWith('DevPlugin-')) {
                    fileExplorerApi.addFolder(selectedPluginId, `${addingNodePath}/${trimmed}`);
                }
            } else {
                newNode.content = '';
                if (selectedPluginId && !selectedPluginId.startsWith('DevPlugin-')) {
                    const formData = new FormData();
                    formData.append('file', new Blob([''], { type: 'text/plain' }), trimmed);
                    fileExplorerApi.uploadFile(selectedPluginId, addingNodePath, formData);
                }
            }
            const updatedTree = addNodeToTree(fileTree, addingNodePath, fileTree.name, newNode);
            updateSelectedPluginState(p => ({ ...p, fileTree: updatedTree }));
        }
        setAddingNodePath(null);
        setAddingNodeType(null);
        setAddingNodeName('');
    };


    return (
        <>
            <LevaOverrides />
            <WorkbenchContainer ref={containerRef}>
                <Title ref={handleRef}>
                <span>Dev Workbench</span>
                {onClose && <CloseBtn onClick={onClose}>×</CloseBtn>}
                </Title>

{!pluginManager ? <div style={{color:'#ff9800', textAlign:'center', padding:'10px'}}>等待进入行星内部以加载插件系统 (按 V 键) ...</div> : null}

                {/* 插件列表 Plugin List */}
                {pluginManager && (
                    <div style={{ padding: '0 10px', marginTop: '10px' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '4px' }}>
                            <div style={{ fontSize: '12px', color: '#888' }}>插件列表</div>
                            <div style={{ display: 'flex', gap: '10px' }}>
                                <FileInput
                                    type="file"
                                    accept=".js,.ts,.txt,.json,.md"
                                    ref={nestedFileInputRef}
                                    onChange={handleNestedFileChange}
                                    style={{ display: 'none' }}
                                />
                                <UploadOutlined 
                                    title="上传插件 (复选文件或上传文件夹)" 
                                    onClick={() => fileInputRef.current?.click()} 
                                    style={{ fontSize: '16px', color: '#4a90e2', cursor: 'pointer' }} 
                                />
                                <input
                                    type="file"
                                    multiple
                                    accept=".js,.ts,.json,.txt,.md"
                                    // @ts-ignore - 非标准属性，浏览器选择文件夹时需要
                                    // webkitdirectory=""
                                    ref={fileInputRef}
                                    onChange={handleFileChange}
                                    style={{ display: 'none' }}
                                />
                                <PlusCircleOutlined 
                                    title="添加插件" 
                                    onClick={() => setShowVisualBuilder(true)} 
                                    style={{ fontSize: '16px', color: '#4a90e2', cursor: 'pointer' }} 
                                />
                            </div>
                        </div>
                        <div style={{ height: '80px', overflowY: 'auto', border: '1px solid #444', borderRadius: '4px', width: '100%' }}>
                            {devPlugins.map(p => {
                                const isSel = selectedPluginId === p.id;
                                const isHov = hoveredPlugin === p.id;
                                return (
                                    <div
                                        key={p.id}
                                        onClick={() => setSelectedPluginId(p.id)}
                                        onMouseEnter={() => setHoveredPlugin(p.id)}
                                        onMouseLeave={() => setHoveredPlugin(null)}
                                        style={{
                                            padding: '4px 8px',
                                            cursor: 'pointer',
                                            backgroundColor: isSel ? 'rgba(255,255,255,0.2)' : (isHov ? 'rgba(255,255,255,0.1)' : 'transparent'),
                                            display: 'flex',
                                            justifyContent: 'space-between',
                                            alignItems: 'center'
                                        }}
                                    >
                                        <div style={{color: '#ccc', fontSize: '13px', display: 'flex', alignItems: 'center', gap: '4px'}}>
                                            {p.id.startsWith('DevPlugin-') && (
                                                <WarningOutlined style={{color: '#faad14'}} title="未保存，页面刷新将丢失" />
                                            )}
                                            {p.name}
                                        </div>
                                        {isHov && (
                                            <div style={{ display: 'flex', gap: '8px' }}>
                                                {p.id.startsWith('DevPlugin-') && (
                                                    <SaveOutlined
                                                        title="保存插件"
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            handleSavePlugin(p.id);
                                                        }}
                                                        style={{ color: '#1890ff' }}
                                                    />
                                                )}
                                                <DeleteOutlined
                                                    title="移除插件"
                                                    onClick={(e: React.MouseEvent) => {
                                                        e.stopPropagation();
                                                        pluginManager.unloadPlugin(p.id);
                                                        const newList = devPlugins.filter(x => x.id !== p.id);
                                                        setDevPlugins(newList);
                                                        if (selectedPluginId === p.id) setSelectedPluginId(newList[0]?.id || null);
                                                    }}
                                                    style={{ color: '#ff4d4f' }}
                                                />
                                            </div>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                )}

                {/* 文件资源管理器 File Explorer */}
                {pluginManager && (
                    <div style={{ padding: '0 10px', marginTop: '10px', width: '100%' }}>
                        <div style={{ fontSize: '12px', color: '#888', marginBottom: '4px' }}>DevWorkbench 资源管理器</div>
                        {fileTree ? (
                            <div style={{ height: '130px', overflowY: 'auto', border: '1px solid #444', borderRadius: '4px', padding: '4px 0', width: '100%' }}>
                                <FileExplorer 
                                    fileTree={fileTree}
                                    selectedNodePath={selectedNodePath}
                                    setSelectedNodePath={setSelectedNodePath}
                                    hoveredNode={hoveredNode}
                                    setHoveredNode={setHoveredNode}
                                    editingNodePath={editingNodePath}
                                    setEditingNodePath={setEditingNodePath}
                                    editingNodeName={editingNodeName}
                                    setEditingNodeName={setEditingNodeName}
                                    addingNodePath={addingNodePath}
                                    setAddingNodePath={setAddingNodePath}
                                    addingNodeType={addingNodeType}
                                    setAddingNodeType={setAddingNodeType}
                                    addingNodeName={addingNodeName}
                                    setAddingNodeName={setAddingNodeName}
                                    nestedFileInputRef={nestedFileInputRef as React.RefObject<HTMLInputElement>}
                                    setNestedFileUploadPath={setNestedFileUploadPath}
                                    onToggleFolder={(path) => {
                                        updateSelectedPluginState(p => ({
                                            ...p,
                                            fileTree: p.fileTree ? toggleFolderOpen(p.fileTree, path, p.fileTree.name) : null
                                        }));
                                    }}
                                    onFileSelect={(path, node) => {
                                        const relativePath = fileTree ? path.replace(`${fileTree.name}/`, '') : node.name;
                                        updateSelectedPluginState(p => ({
                                            ...p,
                                            currentFilename: relativePath,
                                            source: node.content !== undefined ? node.content : p.source
                                        }));
                                        setShowEditor(true);
                                    }}
                                    onReloadPlugin={() => {
                                        if (selectedPlugin) {
                                            const fullSourceFiles = treeToSourceFiles(selectedPlugin.fileTree);
                                            const normalizedSourceFiles = normalizePluginSourceFiles(fullSourceFiles, selectedPlugin.fileTree?.name || selectedPlugin.name);
                                            const entryFileName = resolveEntryFilename(normalizedSourceFiles, selectedPlugin.currentFilename || 'index.js');
                                            const entrySource = normalizedSourceFiles[entryFileName] || selectedPlugin.source;
                                            pluginManager.unloadPlugin(selectedPlugin.id);
                                            pluginManager.loadDynamicPlugin(entrySource, selectedPlugin.id, { id: selectedPlugin.id }, entryFileName, normalizedSourceFiles).then((reloadedInstance: any) => {
                                                if (reloadedInstance) {
                                                    updateSelectedPluginState(p => ({
                                                        ...p,
                                                        source: entrySource,
                                                        currentFilename: entryFileName,
                                                        instance: reloadedInstance
                                                    }));
                                                    messageApi.success('插件已重新加载到场景！');
                                                }
                                            });
                                        }
                                    }}
                                    onUnloadPlugin={() => {
                                        if (selectedPlugin) {
                                            pluginManager.unloadPlugin(selectedPlugin.id);
                                            updateSelectedPluginState(p => ({ ...p, instance: null }));
                                            messageApi.success('插件已从场景卸载！');
                                        }
                                    }}
                                    onRenameSubmit={handleRenameSubmit}
                                    onAddSubmit={handleAddSubmit}
                                    onDeleteSubmit={handleDeleteNode}
                                />
                            </div>
                        ) : (
                            <div style={{ display: 'flex', alignItems: 'center', padding: '15px', color: '#888' }}>
                                <StatusDot $active={!!loadedPluginId} />
                                {loadedPluginId ? loadedPluginId : 'No target loaded'}
                            </div>
                        )}
                    </div>
                )}
            </WorkbenchContainer>
            
            {showVisualBuilder && (
                <VisualBuilder 
                   onClose={() => setShowVisualBuilder(false)}
                   onBuildPlugin={async ({ code, name, language }) => {
                       try {
                            const newId = `DevPlugin-${name}@${Date.now()}`;
                            const ext = language === 'ts' ? 'ts' : 'js';
                            const entryName = `index.${ext}`;
                            const sourceFiles = { [entryName]: code };
                            const reloadedInstance = await pluginManager.loadDynamicPlugin(code, newId, { id: newId }, entryName, sourceFiles);
                            if (reloadedInstance) {
                                const newPlugin: DevPluginData = {
                                    id: newId,
                                    name: name,
                                    source: code,
                                    currentFilename: entryName,
                                    instance: reloadedInstance,
                                    fileTree: {
                                        name: name,
                                        type: 'folder',
                                        isOpen: true,
                                        children: [
                                            {
                                                name: 'manifest.json',
                                                type: 'file',
                                                content: JSON.stringify({
                                                    name: name,
                                                    version: "1.0.0",
                                                    description: "Auto-generated plugin",
                                                    entry: entryName
                                                }, null, 2)
                                            },
                                            {
                                                name: 'publish.json',
                                                type: 'file',
                                                content: JSON.stringify({
                                                    name: name,
                                                    version: "1.0.0",
                                                    description: "Auto-generated plugin",
                                                    totalSupply: 1000,
                                                    mintPer: 1
                                                }, null, 2)
                                            },
                                            {
                                                name: entryName,
                                                type: 'file',
                                                content: code
                                            },
                                            {
                                                name: 'assets',
                                                type: 'folder',
                                                isOpen: true,
                                                children: []
                                            }
                                        ]
                                    }
                                };
                                setDevPlugins(prev => [...prev, newPlugin]);
                                setSelectedPluginId(newId);
                                messageApi.success('Visual Plugin generated and loaded!');
                            }
                       } catch (e: any) {
                            messageApi.error(`Load Error: ${e.message}`);
                       }
                   }}
                />
            )}

            {showEditor && loadedPluginId && (
                <SourceCodeEditor 
                    initialCode={pluginSource}
                    title={currentFilename}
                    sourceFiles={fileTree ? normalizePluginSourceFiles(treeToSourceFiles(fileTree), fileTree.name) : {}}
                    entryFileName={fileTree ? resolveEntryFilename(normalizePluginSourceFiles(treeToSourceFiles(fileTree), fileTree.name), currentFilename || 'index.js') : currentFilename}
                    onClose={() => setShowEditor(false)}
                    onSave={async (newCode) => {
                        try {
                            if (!selectedPlugin || !selectedPlugin.fileTree) {
                                throw new Error('当前插件文件树不存在，无法保存。');
                            }

                            const targetPath = `${selectedPlugin.fileTree.name}/${currentFilename}`;
                            const updatedTree = updateNodeContentInTree(selectedPlugin.fileTree, targetPath, selectedPlugin.fileTree.name, newCode);
                            const normalizedSourceFiles = normalizePluginSourceFiles(treeToSourceFiles(updatedTree), updatedTree.name);
                            const entryFileName = resolveEntryFilename(normalizedSourceFiles, currentFilename || 'index.js');
                            const entrySource = normalizedSourceFiles[entryFileName] || newCode;

                            pluginManager.unloadPlugin(loadedPluginId);
                            const reloadedInstance = await pluginManager.loadDynamicPlugin(entrySource, loadedPluginId, { id: loadedPluginId }, entryFileName, normalizedSourceFiles);
                            if (reloadedInstance) {
                                updateSelectedPluginState(p => ({
                                    ...p,
                                    source: currentFilename === entryFileName ? newCode : entrySource,
                                    currentFilename,
                                    instance: reloadedInstance,
                                    fileTree: updatedTree
                                }));
                                messageApi.success('Code saved and hot-reloaded!');
                            }
                        } catch (err: any) {
                            messageApi.error(`Compile Error: ${err.message}`);
                            console.error(err);
                            throw err;
                        }
                    }}
                />
            )}

            <DevSchemaPanel activePlugin={activePlugin} />
        </>
    );
};

const DevSchemaPanel: React.FC<{ activePlugin: DesktopPlugin | null }> = ({ activePlugin }) => {
    const store = useCreateStore();
    useEffect(() => {
        if (!activePlugin || !activePlugin.getSchema) return;
        const schema = activePlugin.getSchema();
        if (!schema) return;
    }, [activePlugin]);

    if (!activePlugin || !activePlugin.getSchema) return null;

    return (
        <div style={{ position: 'absolute', top: 361, right: 10, zIndex: 1000, width: 320}}>     
            {/* @ts-ignore */}
            <Leva fill titleBar={{ title: '插件参数', filter: false }} hideCopyButton store={store} />
            <SchemaLevaControls key={activePlugin.id} activePlugin={activePlugin!} store={store} />
        </div>
    );
}

const SchemaLevaControls: React.FC<{ activePlugin: DesktopPlugin, store: any }> = ({ activePlugin, store }) => {
    const schema = activePlugin.getSchema ? activePlugin.getSchema() : {};
    
    useControls('Plugin Dev Params', () => {
        const config: any = {};
        for (const [key, def] of Object.entries(schema || {})) {
            const levaKey = `${activePlugin.id}_${key}`;
            config[levaKey] = {
                value: def.default,
                label: def.label || key,
                onChange: (v: any) => {
                    if (activePlugin.onParamChange) {
                         activePlugin.onParamChange(key, v);
                    }
                },
                transient: false
            }
            if (def.type === 'number') {
                if (typeof def.min === 'number') config[levaKey].min = def.min;
                if (typeof def.max === 'number') config[levaKey].max = def.max;
                if (typeof def.step === 'number') config[levaKey].step = def.step;
            }
        }
        return config;
    }, { store }, [activePlugin]);

    return null;
};
