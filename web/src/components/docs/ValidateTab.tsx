'use client';

import { Database, Copy, Check } from 'lucide-react';
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

export default function ValidateTab() {
    return (
        <div className="space-y-8">
            <div>
                <h2 className="text-3xl font-bold text-gray-900 mb-4">Result Validation</h2>
                <div className="bg-purple-50 border border-purple-200 rounded-lg p-6 mb-6">
                    <div className="flex items-start space-x-3">
                        <Database className="h-6 w-6 text-purple-600 mt-1 flex-shrink-0" />
                        <div>
                            <h3 className="font-semibold text-purple-900 mb-2">Quality Assurance</h3>
                            <p className="text-purple-800">
                                The <code className="bg-purple-100 px-1 rounded">validate</code> command tests the accuracy of matching results against known ground truth data.
                                This helps potential users understand that the matching algorithms work correctly.
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Interactive Mode (Recommended)</h3>
                <p className="text-gray-600 mb-3">
                    Interactive mode prompts for configuration files, ground truth data, thresholds, and output preferences.
                </p>
                <CodeBlock command>./cohort-bridge validate</CodeBlock>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Standard Validation</h3>
                <p className="text-gray-600 mb-3">
                    Compare matching results against known ground truth with default thresholds.
                </p>
                <CodeBlock command>./cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/truth.csv</CodeBlock>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Custom Thresholds</h3>
                <p className="text-gray-600 mb-3">
                    Test with stricter matching criteria to see how it affects precision and recall.
                </p>
                <CodeBlock command>./cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/truth.csv -match-threshold 15 -jaccard-threshold 0.8</CodeBlock>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Verbose Analysis</h3>
                <p className="text-gray-600 mb-3">
                    Get detailed breakdown of matches, misses, and false positives for in-depth analysis.
                </p>
                <CodeBlock command>./cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/truth.csv -verbose</CodeBlock>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Automated Validation</h3>
                <p className="text-gray-600 mb-3">
                    Run validation in batch mode without confirmation prompts for automated testing.
                </p>
                <CodeBlock command>./cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/truth.csv -force</CodeBlock>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Complete Parameter Reference</h3>
                <ParameterTable parameters={[
                    { name: '-config1', description: 'Configuration file for dataset 1 (Party A). Contains database path, fields, and matching settings.', default: 'prompted in interactive mode', color: 'bg-purple-100 text-purple-800' },
                    { name: '-config2', description: 'Configuration file for dataset 2 (Party B). Contains database path, fields, and matching settings.', default: 'prompted in interactive mode', color: 'bg-purple-100 text-purple-800' },
                    { name: '-ground-truth', description: 'CSV file with expected matches. Format: two columns with record IDs that should match.', default: 'prompted in interactive mode', color: 'bg-purple-100 text-purple-800' },
                    { name: '-output', description: 'Output CSV file for validation report with precision, recall, F1-score, and detailed match analysis.', default: 'auto-generated from config names', color: 'bg-purple-100 text-purple-800' },
                    { name: '-match-threshold', description: 'Hamming distance threshold for considering records a match. Lower values = stricter matching.', default: '20', color: 'bg-purple-100 text-purple-800' },
                    { name: '-jaccard-threshold', description: 'Minimum Jaccard similarity (0.0-1.0) for considering records a match. Higher values = stricter matching.', default: '0.5', color: 'bg-purple-100 text-purple-800' },
                    { name: '-verbose', description: 'Include detailed analysis with lists of false positives, false negatives, and match breakdowns.', default: 'false', color: 'bg-purple-100 text-purple-800' },
                    { name: '-interactive', description: 'Force interactive mode even when parameters are provided. Useful for reviewing settings.', default: 'false', color: 'bg-purple-100 text-purple-800' },
                    { name: '-force', description: 'Skip confirmation prompts and run automatically. Required for automated testing scripts.', default: 'false', color: 'bg-purple-100 text-purple-800' },
                    { name: '-help', description: 'Show detailed help message with examples and ground truth file format requirements.', default: 'false', color: 'bg-purple-100 text-purple-800' },
                ]} />
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Validation Metrics</h3>
                <p className="text-gray-600 mb-6">
                    The validation report includes comprehensive accuracy metrics:
                </p>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-4">
                        {[
                            { name: 'Precision', color: 'bg-green-100 text-green-800', desc: 'True matches / Total matches found' },
                            { name: 'Recall', color: 'bg-blue-100 text-blue-800', desc: 'True matches / Total actual matches' },
                            { name: 'F1-Score', color: 'bg-purple-100 text-purple-800', desc: 'Harmonic mean of precision and recall' },
                        ].map((metric) => (
                            <div key={metric.name} className="flex items-center space-x-3">
                                <span className={`${metric.color} px-3 py-1 rounded text-sm font-semibold`}>
                                    {metric.name}
                                </span>
                                <span className="text-sm text-gray-600">{metric.desc}</span>
                            </div>
                        ))}
                    </div>
                    <div className="space-y-4">
                        {[
                            { name: 'False Positives', color: 'bg-red-100 text-red-800', desc: 'Incorrect matches found' },
                            { name: 'False Negatives', color: 'bg-orange-100 text-orange-800', desc: 'Actual matches missed' },
                            { name: 'True Positives', color: 'bg-gray-100 text-gray-800', desc: 'Correct matches found' },
                        ].map((metric) => (
                            <div key={metric.name} className="flex items-center space-x-3">
                                <span className={`${metric.color} px-3 py-1 rounded text-sm font-semibold`}>
                                    {metric.name}
                                </span>
                                <span className="text-sm text-gray-600">{metric.desc}</span>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
} 