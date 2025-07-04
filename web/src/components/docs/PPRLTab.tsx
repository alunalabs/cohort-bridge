'use client';

import { Network, Copy, Check } from 'lucide-react';
import { useState } from 'react';

const CodeBlock = ({ children, command }: { children: string; command?: boolean }) => {
    const [copied, setCopied] = useState(false);

    const copyCommand = async (text: string) => {
        try {
            await navigator.clipboard.writeText(text);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy: ', err);
        }
    };

    return (
        <div className="bg-slate-50 border border-slate-200 rounded p-3 relative overflow-hidden">
            <div className="overflow-x-auto">
                <code className="text-slate-700 text-xs font-mono whitespace-pre block">
                    {children}
                </code>
            </div>
            {command && (
                <button
                    onClick={() => copyCommand(children)}
                    className="absolute top-2 right-2 p-1 hover:bg-slate-200 rounded transition-colors"
                    title="Copy to clipboard"
                >
                    {copied ? (
                        <Check className="h-3 w-3 text-green-600" />
                    ) : (
                        <Copy className="h-3 w-3 text-gray-500" />
                    )}
                </button>
            )}
        </div>
    );
};

const ParameterTable = ({ parameters }: {
    parameters: Array<{
        name: string;
        description: string;
        default: string;
        color?: string;
    }>
}) => (
    <div className="overflow-x-auto">
        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden shadow-sm">
            <table className="w-full text-sm">
                <thead>
                    <tr className="bg-gray-50 border-b border-gray-200">
                        <th className="px-6 py-3 text-left font-semibold text-gray-900">Parameter</th>
                        <th className="px-6 py-3 text-left font-semibold text-gray-900">Description</th>
                        <th className="px-6 py-3 text-left font-semibold text-gray-900">Default</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                    {parameters.map((param, index) => (
                        <tr key={index} className="hover:bg-gray-50 transition-colors">
                            <td className="px-6 py-3">
                                <code className={`${param.color || 'bg-gray-100 text-gray-800'} px-2 py-1 rounded text-xs font-mono`}>
                                    {param.name}
                                </code>
                            </td>
                            <td className="px-6 py-3 text-gray-700">{param.description}</td>
                            <td className="px-6 py-3 text-gray-600 font-mono text-xs">{param.default}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    </div>
);

export default function PPRLTab() {
    return (
        <div className="space-y-8">
            <div>
                <h2 className="text-3xl font-bold text-gray-900 mb-4">Peer-to-Peer Privacy-Preserving Record Linkage</h2>
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-6">
                    <div className="flex items-start space-x-3">
                        <Network className="h-6 w-6 text-blue-600 mt-1 flex-shrink-0" />
                        <div>
                            <h3 className="font-semibold text-blue-900 mb-2">Main Command</h3>
                            <p className="text-blue-800">
                                The <code className="bg-blue-100 px-1 rounded">pprl</code> command orchestrates a complete 7-step peer-to-peer workflow
                                for finding intersections between two parties' databases using zero-knowledge protocols.
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Interactive Mode (Recommended)</h3>
                <p className="text-gray-600 mb-3">
                    Interactive mode guides you through configuration and shows prompts for confirmation at each step.
                    This is the safest way to run PPRL operations.
                </p>
                <CodeBlock command>./cohort-bridge pprl</CodeBlock>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Configuration File Mode</h3>
                <p className="text-gray-600 mb-3">
                    Use a pre-created configuration file. Still shows confirmation prompts unless combined with -force.
                </p>
                <CodeBlock command>./cohort-bridge pprl -config config.yaml</CodeBlock>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Automated Mode (For Scripts)</h3>
                <p className="text-gray-600 mb-3">
                    Use -force to skip all confirmation prompts. Essential when calling from other programs or scripts.
                </p>
                <CodeBlock command>./cohort-bridge pprl -config config.yaml -force</CodeBlock>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Allow Multiple Matches</h3>
                <p className="text-gray-600 mb-3">
                    By default, CohortBridge enforces 1:1 matching (each record matches at most one other record).
                    Use -allow-duplicates to enable 1:many matching.
                </p>
                <CodeBlock command>./cohort-bridge pprl -config config.yaml -allow-duplicates</CodeBlock>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Complete Parameter Reference</h3>
                <ParameterTable parameters={[
                    { name: '-config', description: 'Path to YAML configuration file containing peer settings, thresholds, and database configuration', default: 'none (required)', color: 'bg-blue-100 text-blue-800' },
                    { name: '-interactive', description: 'Force interactive mode even when config file is provided. Useful for reviewing settings.', default: 'false', color: 'bg-blue-100 text-blue-800' },
                    { name: '-force', description: 'Skip all confirmation prompts and run automatically. Required when calling from scripts or other programs.', default: 'false', color: 'bg-blue-100 text-blue-800' },
                    { name: '-allow-duplicates', description: 'Allow 1:many matching where one record can match multiple peer records. Default is 1:1 matching only.', default: 'false', color: 'bg-blue-100 text-blue-800' },
                    { name: '-help', description: 'Show detailed help message with examples and configuration requirements', default: 'false', color: 'bg-blue-100 text-blue-800' },
                ]} />
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Workflow</h3>
                <p className="text-gray-600 mb-6">
                    The PPRL command orchestrates a complete privacy-preserving workflow:
                </p>
                <div className="space-y-4">
                    {[
                        { step: 1, title: 'Configuration Loading', desc: 'Reads your config file and validates settings' },
                        { step: 2, title: 'Data Tokenization', desc: 'Converts your PHI data into privacy-preserving tokens' },
                        { step: 3, title: 'Peer Connection', desc: 'Establishes secure connection with the other party' },
                        { step: 4, title: 'Token Exchange', desc: 'Securely exchanges tokenized data' },
                        { step: 5, title: 'Intersection Computing', desc: 'Finds matches using zero-knowledge protocols' },
                        { step: 6, title: 'Result Validation', desc: 'Compares results between parties to ensure accuracy' },
                    ].map((item) => (
                        <div key={item.step} className="flex items-start space-x-4">
                            <span className="bg-blue-100 text-blue-800 rounded-full w-8 h-8 flex items-center justify-center text-sm font-semibold flex-shrink-0">
                                {item.step}
                            </span>
                            <div>
                                <h4 className="font-semibold text-gray-900">{item.title}</h4>
                                <p className="text-gray-600">{item.desc}</p>
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Connection Details</h3>
                <p className="text-gray-600 mb-4">
                    CohortBridge automatically handles the connection process. One party becomes the server, the other becomes the client:
                </p>

                <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-4">
                    <h4 className="font-semibold text-gray-900 mb-2">How Connection Works</h4>
                    <ol className="list-decimal list-inside text-gray-600 text-sm space-y-1">
                        <li>CohortBridge first tries to connect as a client to the peer address</li>
                        <li>If that fails, it starts listening as a server on the configured port</li>
                        <li>The other party will connect as a client to complete the connection</li>
                        <li>Once connected, both parties exchange their tokenized data</li>
                        <li>Results are computed and compared for consistency</li>
                    </ol>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                        <h4 className="font-semibold text-blue-900 mb-2">Network Requirements</h4>
                        <ul className="list-disc list-inside text-blue-800 text-sm space-y-1">
                            <li>Direct TCP connection between parties</li>
                            <li>Firewall allows traffic on chosen ports</li>
                            <li>Both parties can reach each other's IP</li>
                            <li>Network supports bidirectional communication</li>
                        </ul>
                    </div>

                    <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                        <h4 className="font-semibold text-green-900 mb-2">Security Features</h4>
                        <ul className="list-disc list-inside text-green-800 text-sm space-y-1">
                            <li>All communication is encrypted</li>
                            <li>Only tokens are exchanged, never raw data</li>
                            <li>Zero-knowledge matching protocols</li>
                            <li>Results validation prevents tampering</li>
                        </ul>
                    </div>
                </div>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Troubleshooting</h3>
                <div className="space-y-4">
                    <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                        <h4 className="font-semibold text-red-900 mb-2">Connection Failed</h4>
                        <p className="text-red-800 text-sm mb-2">If you see "Failed to establish peer connection":</p>
                        <ul className="list-disc list-inside text-red-800 text-sm space-y-1">
                            <li>Check both parties have correct IP addresses configured</li>
                            <li>Verify firewall allows traffic on the specified ports</li>
                            <li>Test basic connectivity with <code>ping [partner_ip]</code></li>
                            <li>Try switching which party starts first</li>
                            <li>Check for network proxy or VPN interference</li>
                        </ul>
                    </div>

                    <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                        <h4 className="font-semibold text-yellow-900 mb-2">Results Don't Match</h4>
                        <p className="text-yellow-800 text-sm mb-2">If intersection results differ between parties:</p>
                        <ul className="list-disc list-inside text-yellow-800 text-sm space-y-1">
                            <li>Ensure both parties use identical field configurations</li>
                            <li>Check that matching thresholds are the same</li>
                            <li>Verify the same normalization rules are applied</li>
                            <li>Review the diff file in the output directory</li>
                            <li>Compare configuration files between parties</li>
                        </ul>
                    </div>

                    <div className="bg-purple-50 border border-purple-200 rounded-lg p-4">
                        <h4 className="font-semibold text-purple-900 mb-2">Performance Issues</h4>
                        <p className="text-purple-800 text-sm mb-2">For slow processing or timeouts:</p>
                        <ul className="list-disc list-inside text-purple-800 text-sm space-y-1">
                            <li>Start with smaller datasets for testing</li>
                            <li>Monitor memory usage during large dataset processing</li>
                            <li>Check network bandwidth between parties</li>
                            <li>Consider using pre-tokenized data for faster processing</li>
                            <li>Adjust timeout values in configuration if needed</li>
                        </ul>
                    </div>
                </div>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Output Files</h3>
                <p className="text-gray-600 mb-4">
                    After successful completion, check the <code>out/</code> directory for results:
                </p>
                <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                    <div className="space-y-2 font-mono text-sm">
                        <div className="flex justify-between">
                            <span className="text-blue-600">intersection_results.json</span>
                            <span className="text-gray-500">Matched record pairs</span>
                        </div>
                        <div className="flex justify-between">
                            <span className="text-green-600">tokenized_data.csv</span>
                            <span className="text-gray-500">Your processed tokens</span>
                        </div>
                        <div className="flex justify-between">
                            <span className="text-purple-600">cohort-bridge.log</span>
                            <span className="text-gray-500">Detailed process log</span>
                        </div>
                        <div className="flex justify-between">
                            <span className="text-red-600">intersection_diff.json</span>
                            <span className="text-gray-500">Only if results differ</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
} 