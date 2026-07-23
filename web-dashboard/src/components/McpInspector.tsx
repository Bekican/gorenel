import React, { useState, useMemo } from 'react';
import { Search, Bot, Cpu, Terminal, Clock, ShieldCheck, AlertCircle, ChevronRight, Database } from 'lucide-react';
import { format } from 'date-fns';
import type { CapturedRequest } from '../api/client';

interface McpInspectorProps {
  history: CapturedRequest[];
  activeSubdomains: string[];
}

interface McpTool {
  name: string;
  description?: string;
  inputSchema?: {
    type: string;
    properties?: Record<string, any>;
    required?: string[];
  };
}

interface ToolCallRecord {
  id: string;
  timestamp: string;
  toolName: string;
  arguments: any;
  result: any;
  error?: string;
  durationMs: number;
}

const decodeBase64Utf8 = (base64Str: string): string => {
  try {
    const binaryString = atob(base64Str);
    const len = binaryString.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) {
      bytes[i] = binaryString.charCodeAt(i);
    }
    return new TextDecoder('utf-8').decode(bytes);
  } catch (e) {
    return atob(base64Str);
  }
};

export const McpInspector: React.FC<McpInspectorProps> = ({ history, activeSubdomains }) => {
  const [selectedSubdomain, setSelectedSubdomain] = useState<string>(activeSubdomains[0] || '');
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedCallId, setSelectedCallId] = useState<string | null>(null);
  const [activeSubTab, setActiveSubTab] = useState<'calls' | 'tools' | 'raw'>('calls');

  // Filter history for the selected subdomain and JSON-RPC payloads
  const mcpRequests = useMemo(() => {
    if (!selectedSubdomain) return [];
    return history.filter(req => {
      if (req.subdomain !== selectedSubdomain) return false;
      
      // Check if it's SSE endpoint or POST message endpoint
      if (req.path === '/sse' || req.path.startsWith('/message')) return true;

      // Try to check if request body contains JSON-RPC
      try {
        const bodyStr = decodeBase64Utf8(req.req_body);
        return bodyStr.includes('"jsonrpc"') || bodyStr.includes('"method"');
      } catch {
        return false;
      }
    });
  }, [history, selectedSubdomain]);

  // Extract tool declarations from the latest "tools/list" response
  const registeredTools = useMemo(() => {
    // Traverse backwards to find the latest tools/list response
    for (let i = mcpRequests.length - 1; i >= 0; i--) {
      const req = mcpRequests[i];
      try {
        const reqBody = JSON.parse(decodeBase64Utf8(req.req_body));
        if (reqBody.method === 'tools/list') {
          const respBody = JSON.parse(decodeBase64Utf8(req.resp_body));
          if (respBody.result && Array.isArray(respBody.result.tools)) {
            return respBody.result.tools as McpTool[];
          }
        }
      } catch {
        // Skip malformed
      }
    }
    return [] as McpTool[];
  }, [mcpRequests]);

  // Extract all "tools/call" RPC operations
  const toolCalls = useMemo(() => {
    const list: ToolCallRecord[] = [];
    mcpRequests.forEach(req => {
      try {
        const reqBody = JSON.parse(decodeBase64Utf8(req.req_body));
        if (reqBody.method === 'tools/call') {
          const respBody = JSON.parse(decodeBase64Utf8(req.resp_body));
          
          let result = null;
          let error = undefined;

          if (respBody.error) {
            error = respBody.error.message || JSON.stringify(respBody.error);
          } else if (respBody.result) {
            result = respBody.result;
          }

          list.push({
            id: req.id,
            timestamp: req.timestamp,
            toolName: reqBody.params?.name || 'unknown',
            arguments: reqBody.params?.arguments || {},
            result: result,
            error: error,
            durationMs: Math.round(req.duration / 1000000), // convert ns to ms
          });
        }
      } catch {
        // Skip non-json or malformed
      }
    });
    // Return newest first
    return list.reverse();
  }, [mcpRequests]);

  const selectedCall = useMemo(() => {
    return toolCalls.find(c => c.id === selectedCallId) || null;
  }, [toolCalls, selectedCallId]);

  const filteredCalls = useMemo(() => {
    return toolCalls.filter(c => 
      c.toolName.toLowerCase().includes(searchTerm.toLowerCase()) ||
      JSON.stringify(c.arguments).toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [toolCalls, searchTerm]);

  return (
    <div className="flex flex-col h-[calc(100dvh-220px)] sm:h-[calc(100vh-220px)] min-h-[520px] overflow-hidden bg-black/20 border border-white/[0.04] rounded-2xl">
      {/* Subdomain selector & Top Controls */}
      <div className="flex flex-wrap items-center justify-between border-b border-white/[0.06] p-4 gap-3 bg-white/[0.01]">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-purple-500/10 border border-purple-500/20 flex items-center justify-center text-purple-400">
            <Bot className="w-4 h-4" />
          </div>
          <div>
            <h2 className="text-sm font-semibold text-white/95">MCP Telemetry Inspector</h2>
            <p className="text-[11px] text-white/40">Exposing Model Context Protocol servers securely</p>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <span className="text-xs text-white/40 font-medium">Select Tunnel:</span>
          <select
            value={selectedSubdomain}
            onChange={(e) => {
              setSelectedSubdomain(e.target.value);
              setSelectedCallId(null);
            }}
            className="px-3 py-1.5 rounded-lg bg-white/[0.03] border border-white/[0.08] text-xs text-white/80 focus:border-purple-500 focus:outline-none"
          >
            {activeSubdomains.length === 0 ? (
              <option value="">No Active MCP Tunnels</option>
            ) : (
              activeSubdomains.map(sub => (
                <option key={sub} value={sub}>{sub}.gorenel.site</option>
              ))
            )}
          </select>

          {/* Sub-tab toggles */}
          <div className="flex items-center rounded-lg bg-white/[0.03] border border-white/[0.08] p-0.5 text-xs">
            <button
              onClick={() => setActiveSubTab('calls')}
              className={`px-3 py-1 rounded-md transition-colors ${activeSubTab === 'calls' ? 'bg-purple-500/25 text-purple-300 font-semibold' : 'text-white/60 hover:text-white/80'}`}
            >
              Tool Calls
            </button>
            <button
              onClick={() => setActiveSubTab('tools')}
              className={`px-3 py-1 rounded-md transition-colors ${activeSubTab === 'tools' ? 'bg-purple-500/25 text-purple-300 font-semibold' : 'text-white/60 hover:text-white/80'}`}
            >
              Tool Catalog ({registeredTools.length})
            </button>
            <button
              onClick={() => setActiveSubTab('raw')}
              className={`px-3 py-1 rounded-md transition-colors ${activeSubTab === 'raw' ? 'bg-purple-500/25 text-purple-300 font-semibold' : 'text-white/60 hover:text-white/80'}`}
            >
              Raw RPC Stream
            </button>
          </div>
        </div>
      </div>

      {activeSubdomains.length === 0 ? (
        <div className="flex-1 flex flex-col items-center justify-center p-8 text-center">
          <Cpu className="w-12 h-12 text-white/10 mb-3 animate-pulse" />
          <h3 className="text-sm font-semibold text-white/85">No Active MCP Tunnel Found</h3>
          <p className="text-xs text-white/40 mt-1 max-w-[280px]">
            Run <code className="px-1.5 py-0.5 rounded bg-white/[0.04] text-purple-300">gorenel mcp --command "node server.js"</code> in your CLI to expose an MCP server.
          </p>
        </div>
      ) : (
        <div className="flex-1 flex overflow-hidden">
          {activeSubTab === 'calls' && (
            <>
              {/* Tool Calls List */}
              <div className="w-full md:w-1/2 border-r border-white/[0.06] flex flex-col overflow-hidden">
                <div className="p-3 border-b border-white/[0.06] bg-white/[0.01]">
                  <div className="relative">
                    <Search className="absolute left-3 top-2.5 w-3.5 h-3.5 text-white/30" />
                    <input
                      type="text"
                      placeholder="Search tool name or args..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="w-full bg-white/[0.03] border border-white/[0.08] rounded-lg pl-9 pr-4 py-2 text-xs text-white/80 placeholder-white/30 focus:outline-none focus:border-purple-500"
                    />
                  </div>
                </div>

                <div className="flex-1 overflow-y-auto divide-y divide-white/[0.03]">
                  {filteredCalls.length === 0 ? (
                    <div className="p-8 text-center text-white/30 text-xs">
                      No tool calls captured yet.
                    </div>
                  ) : (
                    filteredCalls.map((call) => (
                      <div
                        key={call.id}
                        onClick={() => setSelectedCallId(call.id)}
                        className={`p-4 cursor-pointer transition-colors text-left ${selectedCallId === call.id ? 'bg-purple-500/10' : 'hover:bg-white/[0.02]'}`}
                      >
                        <div className="flex items-center justify-between mb-1.5">
                          <span className="font-semibold text-xs text-white/95 flex items-center gap-1.5">
                            <Terminal className="w-3.5 h-3.5 text-purple-400" />
                            {call.toolName}
                          </span>
                          <span className="text-[10px] text-white/40 flex items-center gap-1">
                            <Clock className="w-2.5 h-2.5" />
                            {call.durationMs}ms
                          </span>
                        </div>
                        <div className="text-[11px] text-white/60 truncate font-mono bg-black/25 px-2 py-1 rounded border border-white/[0.02]">
                          Args: {JSON.stringify(call.arguments)}
                        </div>
                        <div className="flex items-center justify-between mt-2">
                          <span className="text-[9px] text-white/30">
                            {format(new Date(call.timestamp), 'HH:mm:ss.SSS')}
                          </span>
                          {call.error ? (
                            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-rose-500/10 border border-rose-500/20 text-[9px] text-rose-400">
                              <AlertCircle className="w-2.5 h-2.5" /> Failed
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20 text-[9px] text-emerald-400">
                              <ShieldCheck className="w-2.5 h-2.5" /> Success
                            </span>
                          )}
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </div>

              {/* Call Details Panel */}
              <div className="hidden md:flex flex-1 flex-col overflow-hidden bg-white/[0.005]">
                {selectedCall ? (
                  <div className="flex-1 flex flex-col overflow-hidden p-5 space-y-4">
                    <div className="flex items-center justify-between border-b border-white/[0.06] pb-3">
                      <div>
                        <h3 className="font-bold text-sm text-white/95">{selectedCall.toolName}</h3>
                        <p className="text-[10px] text-white/30">Latency: {selectedCall.durationMs}ms | Timestamp: {selectedCall.timestamp}</p>
                      </div>
                      <div>
                        {selectedCall.error ? (
                          <span className="px-2 py-1 rounded bg-rose-500/10 border border-rose-500/20 text-xs text-rose-400 font-semibold">
                            FAILED
                          </span>
                        ) : (
                          <span className="px-2 py-1 rounded bg-emerald-500/10 border border-emerald-500/20 text-xs text-emerald-400 font-semibold">
                            SUCCESS
                          </span>
                        )}
                      </div>
                    </div>

                    <div className="flex-1 overflow-y-auto space-y-4 text-left">
                      {/* Arguments */}
                      <div>
                        <h4 className="text-xs font-semibold text-white/60 mb-1.5 flex items-center gap-1.5">
                          <ChevronRight className="w-3.5 h-3.5 text-purple-400" /> Input Arguments
                        </h4>
                        <pre className="p-3.5 rounded-xl border border-white/[0.04] bg-black/30 text-[11px] font-mono text-white/80 overflow-x-auto">
                          {JSON.stringify(selectedCall.arguments, null, 2)}
                        </pre>
                      </div>

                      {/* Result / Content */}
                      <div>
                        <h4 className="text-xs font-semibold text-white/60 mb-1.5 flex items-center gap-1.5">
                          <ChevronRight className="w-3.5 h-3.5 text-purple-400" /> Response Output
                        </h4>
                        {selectedCall.error ? (
                          <pre className="p-3.5 rounded-xl border border-rose-500/10 bg-rose-500/[0.03] text-[11px] font-mono text-rose-300 overflow-x-auto">
                            {selectedCall.error}
                          </pre>
                        ) : (
                          <pre className="p-3.5 rounded-xl border border-white/[0.04] bg-black/30 text-[11px] font-mono text-emerald-300/90 overflow-x-auto">
                            {JSON.stringify(selectedCall.result, null, 2)}
                          </pre>
                        )}
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="flex-1 flex flex-col items-center justify-center text-white/20 text-xs">
                    Select a tool call from the list to view its inputs and outputs.
                  </div>
                )}
              </div>
            </>
          )}

          {activeSubTab === 'tools' && (
            <div className="flex-1 overflow-y-auto p-6 text-left">
              <h3 className="text-sm font-bold text-white/90 mb-4 flex items-center gap-2">
                <Database className="w-4 h-4 text-purple-400" />
                Available Tools Registered on local MCP Server ({registeredTools.length})
              </h3>
              {registeredTools.length === 0 ? (
                <div className="p-12 text-center border border-dashed border-white/[0.06] rounded-xl text-white/30 text-xs">
                  No tools metadata captured yet. The AI agent must query `tools/list` first.
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {registeredTools.map((tool) => (
                    <div key={tool.name} className="p-4 rounded-xl border border-white/[0.06] bg-white/[0.01] hover:border-white/[0.1] transition-colors">
                      <h4 className="font-semibold text-xs text-purple-300 font-mono mb-1">{tool.name}</h4>
                      <p className="text-[11px] text-white/50 mb-3">{tool.description || 'No description provided.'}</p>
                      
                      {tool.inputSchema && tool.inputSchema.properties && (
                        <div className="mt-2 bg-black/20 p-2.5 rounded-lg border border-white/[0.02]">
                          <span className="text-[9px] text-white/40 block mb-1 font-semibold">Schema Parameters:</span>
                          <div className="space-y-1">
                            {Object.entries(tool.inputSchema.properties).map(([name, val]: [string, any]) => (
                              <div key={name} className="flex items-start justify-between text-[10px]">
                                <span className="font-mono text-white/70">
                                  {name}
                                  {tool.inputSchema?.required?.includes(name) && <span className="text-rose-400">*</span>}
                                </span>
                                <span className="text-white/40 text-[9px] font-mono">({val.type || 'any'})</span>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeSubTab === 'raw' && (
            <div className="flex-1 flex flex-col overflow-hidden">
              <div className="flex-1 p-4 overflow-y-auto font-mono text-[10px] text-white/60 bg-black/40 space-y-2 text-left">
                {mcpRequests.map((req) => {
                  let reqParsed = {};
                  let respParsed = {};
                  try {
                    reqParsed = JSON.parse(decodeBase64Utf8(req.req_body));
                    respParsed = JSON.parse(decodeBase64Utf8(req.resp_body));
                  } catch {
                    // Fail gracefully
                  }

                  return (
                    <div key={req.id} className="border-b border-white/[0.03] pb-3 mb-3">
                      <div className="flex items-center gap-3 text-white/30 text-[9px] mb-1.5">
                        <span>{format(new Date(req.timestamp), 'yyyy-MM-dd HH:mm:ss.SSS')}</span>
                        <span className="px-1.5 py-0.5 rounded bg-white/[0.03] text-white/40">{req.method} {req.path}</span>
                        <span>Status: {req.status_code}</span>
                      </div>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                        <div className="bg-white/[0.01] border border-white/[0.04] p-2.5 rounded-lg">
                          <span className="text-purple-400/80 font-bold block mb-1">RPC Request:</span>
                          <pre className="overflow-x-auto text-[9px] max-h-40">{JSON.stringify(reqParsed, null, 2)}</pre>
                        </div>
                        <div className="bg-white/[0.01] border border-white/[0.04] p-2.5 rounded-lg">
                          <span className="text-emerald-400/80 font-bold block mb-1">RPC Response:</span>
                          <pre className="overflow-x-auto text-[9px] max-h-40">{JSON.stringify(respParsed, null, 2)}</pre>
                        </div>
                      </div>
                    </div>
                  );
                })}
                {mcpRequests.length === 0 && (
                  <div className="p-8 text-center text-white/30">
                    No RPC streams captured on this tunnel yet.
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
