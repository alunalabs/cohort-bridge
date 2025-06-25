'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowLeft, Github, Download, Terminal, FileText, Copy, Check, ExternalLink, ChevronRight } from 'lucide-react';

export default function GetStartedPage() {
    const router = useRouter();
    const [copiedStep, setCopiedStep] = useState<string | null>(null);
    const [selectedOS, setSelectedOS] = useState<'windows' | 'macos' | 'linux'>('windows');

    const copyToClipboard = (text: string, stepId: string) => {
        navigator.clipboard.writeText(text);
        setCopiedStep(stepId);
        setTimeout(() => setCopiedStep(null), 2000);
    };

    const installCommands = {
        windows: {
            download: 'Invoke-WebRequest -Uri "https://github.com/alunalabs/cohort-bridge/releases/latest/download/cohort-bridge-windows.zip" -OutFile "cohort-bridge.zip"',
            extract: 'Expand-Archive -Path "cohort-bridge.zip" -DestinationPath "C:\\Program Files\\CohortBridge"',
            path: '$env:PATH += ";C:\\Program Files\\CohortBridge"'
        },
        macos: {
            download: 'curl -L "https://github.com/alunalabs/cohort-bridge/releases/latest/download/cohort-bridge-macos.tar.gz" -o cohort-bridge.tar.gz',
            extract: 'tar -xzf cohort-bridge.tar.gz -C /usr/local/bin/',
            path: 'echo \'export PATH="/usr/local/bin:$PATH"\' >> ~/.zshrc && source ~/.zshrc'
        },
        linux: {
            download: 'wget https://github.com/alunalabs/cohort-bridge/releases/latest/download/cohort-bridge-linux.tar.gz',
            extract: 'tar -xzf cohort-bridge-linux.tar.gz -C /usr/local/bin/',
            path: 'echo \'export PATH="/usr/local/bin:$PATH"\' >> ~/.bashrc && source ~/.bashrc'
        }
    };

    const configTemplates = [
        {
            name: 'Basic CSV Configuration',
            description: 'Simple setup for CSV file record linkage',
            command: 'cohort-bridge init --template basic --output config.yaml'
        },
        {
            name: 'PostgreSQL Configuration',
            description: 'Database connection with secure matching',
            command: 'cohort-bridge init --template postgres --output config-postgres.yaml'
        },
        {
            name: 'Secure Multi-Party',
            description: 'Enhanced security for sensitive data',
            command: 'cohort-bridge init --template secure --output config-secure.yaml'
        },
        {
            name: 'Tokenized Workflow',
            description: 'Pre-tokenized data processing',
            command: 'cohort-bridge init --template tokenized --output config-tokenized.yaml'
        }
    ];

    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-100">
            {/* Header */}
            <header className="bg-white shadow-sm border-b border-slate-200">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex justify-between items-center py-6">
                        <div className="flex items-center space-x-4">
                            <button
                                onClick={() => router.push('/')}
                                className="flex items-center space-x-2 text-slate-600 hover:text-slate-900 transition-colors cursor-pointer"
                            >
                                <ArrowLeft className="h-5 w-5" />
                                <span>Back to Home</span>
                            </button>
                        </div>
                        <div className="flex items-center space-x-3">
                            <div className="w-8 h-8 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
                                <Terminal className="h-4 w-4 text-white" />
                            </div>
                            <div>
                                <h1 className="text-lg font-bold text-slate-900">CohortBridge</h1>
                                <p className="text-xs text-slate-600">Get Started</p>
                            </div>
                        </div>
                    </div>
                </div>
            </header>

            <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
                {/* Hero Section */}
                <div className="text-center mb-16">
                    <h2 className="text-4xl font-bold text-slate-900 mb-6">
                        Get Started with CohortBridge
                    </h2>
                    <p className="text-xl text-slate-600 max-w-3xl mx-auto leading-relaxed">
                        Follow these simple steps to install the CLI tool and create your first configuration.
                        You'll be up and running in just a few minutes!
                    </p>
                </div>

                {/* Steps */}
                <div className="space-y-8">
                    {/* Step 1: Visit GitHub */}
                    <div className="bg-white rounded-xl shadow-lg border border-slate-200 p-8">
                        <div className="flex items-start space-x-4">
                            <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center flex-shrink-0">
                                <span className="text-xl font-bold text-blue-600">1</span>
                            </div>
                            <div className="flex-1">
                                <h3 className="text-xl font-semibold text-slate-900 mb-3">
                                    Visit the GitHub Repository
                                </h3>
                                <p className="text-slate-600 mb-4">
                                    Head to our GitHub repository to access the latest releases, documentation, and source code.
                                </p>
                                <button
                                    onClick={() => window.open('https://github.com/alunalabs/cohort-bridge', '_blank')}
                                    className="inline-flex items-center space-x-2 bg-slate-900 text-white px-6 py-3 rounded-lg hover:bg-slate-800 transition-colors cursor-pointer"
                                >
                                    <Github className="h-5 w-5" />
                                    <span>Open GitHub Repository</span>
                                    <ExternalLink className="h-4 w-4" />
                                </button>
                            </div>
                        </div>
                    </div>

                    {/* Step 2: Install CLI Tool */}
                    <div className="bg-white rounded-xl shadow-lg border border-slate-200 p-8">
                        <div className="flex items-start space-x-4">
                            <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center flex-shrink-0">
                                <span className="text-xl font-bold text-green-600">2</span>
                            </div>
                            <div className="flex-1">
                                <h3 className="text-xl font-semibold text-slate-900 mb-3">
                                    Install the CLI Tool
                                </h3>
                                <p className="text-slate-600 mb-4">
                                    Download and install the CohortBridge CLI tool for your operating system.
                                </p>

                                {/* OS Selection */}
                                <div className="flex space-x-2 mb-6">
                                    {(['windows', 'macos', 'linux'] as const).map((os) => (
                                        <button
                                            key={os}
                                            onClick={() => setSelectedOS(os)}
                                            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors cursor-pointer ${selectedOS === os
                                                    ? 'bg-blue-600 text-white'
                                                    : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
                                                }`}
                                        >
                                            {os === 'macos' ? 'macOS' : os.charAt(0).toUpperCase() + os.slice(1)}
                                        </button>
                                    ))}
                                </div>

                                {/* Installation Commands */}
                                <div className="space-y-4">
                                    {/* Download */}
                                    <div>
                                        <h4 className="text-sm font-semibold text-slate-900 mb-2">Download</h4>
                                        <div className="bg-slate-900 rounded-lg p-4 relative">
                                            <code className="text-green-400 text-sm font-mono">
                                                {installCommands[selectedOS].download}
                                            </code>
                                            <button
                                                onClick={() => copyToClipboard(installCommands[selectedOS].download, `download-${selectedOS}`)}
                                                className="absolute top-2 right-2 p-2 hover:bg-slate-700 rounded transition-colors cursor-pointer"
                                            >
                                                {copiedStep === `download-${selectedOS}` ? (
                                                    <Check className="h-4 w-4 text-green-400" />
                                                ) : (
                                                    <Copy className="h-4 w-4 text-slate-400" />
                                                )}
                                            </button>
                                        </div>
                                    </div>

                                    {/* Extract */}
                                    <div>
                                        <h4 className="text-sm font-semibold text-slate-900 mb-2">Extract & Install</h4>
                                        <div className="bg-slate-900 rounded-lg p-4 relative">
                                            <code className="text-green-400 text-sm font-mono">
                                                {installCommands[selectedOS].extract}
                                            </code>
                                            <button
                                                onClick={() => copyToClipboard(installCommands[selectedOS].extract, `extract-${selectedOS}`)}
                                                className="absolute top-2 right-2 p-2 hover:bg-slate-700 rounded transition-colors cursor-pointer"
                                            >
                                                {copiedStep === `extract-${selectedOS}` ? (
                                                    <Check className="h-4 w-4 text-green-400" />
                                                ) : (
                                                    <Copy className="h-4 w-4 text-slate-400" />
                                                )}
                                            </button>
                                        </div>
                                    </div>

                                    {/* Add to PATH */}
                                    <div>
                                        <h4 className="text-sm font-semibold text-slate-900 mb-2">Add to PATH (Recommended)</h4>
                                        <div className="bg-slate-900 rounded-lg p-4 relative">
                                            <code className="text-green-400 text-sm font-mono">
                                                {installCommands[selectedOS].path}
                                            </code>
                                            <button
                                                onClick={() => copyToClipboard(installCommands[selectedOS].path, `path-${selectedOS}`)}
                                                className="absolute top-2 right-2 p-2 hover:bg-slate-700 rounded transition-colors cursor-pointer"
                                            >
                                                {copiedStep === `path-${selectedOS}` ? (
                                                    <Check className="h-4 w-4 text-green-400" />
                                                ) : (
                                                    <Copy className="h-4 w-4 text-slate-400" />
                                                )}
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Step 3: Create Configuration */}
                    <div className="bg-white rounded-xl shadow-lg border border-slate-200 p-8">
                        <div className="flex items-start space-x-4">
                            <div className="w-12 h-12 bg-purple-100 rounded-full flex items-center justify-center flex-shrink-0">
                                <span className="text-xl font-bold text-purple-600">3</span>
                            </div>
                            <div className="flex-1">
                                <h3 className="text-xl font-semibold text-slate-900 mb-3">
                                    Create Your Configuration
                                </h3>
                                <p className="text-slate-600 mb-6">
                                    Choose from our pre-built templates or create a configuration from scratch.
                                </p>

                                {/* Templates */}
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
                                    {configTemplates.map((template, index) => (
                                        <div
                                            key={index}
                                            className="border border-slate-200 rounded-lg p-4 hover:border-slate-300 transition-colors"
                                        >
                                            <h4 className="font-semibold text-slate-900 mb-1">{template.name}</h4>
                                            <p className="text-sm text-slate-600 mb-3">{template.description}</p>
                                            <div className="bg-slate-900 rounded p-2 relative">
                                                <code className="text-green-400 text-xs font-mono">
                                                    {template.command}
                                                </code>
                                                <button
                                                    onClick={() => copyToClipboard(template.command, `template-${index}`)}
                                                    className="absolute top-1 right-1 p-1 hover:bg-slate-700 rounded transition-colors cursor-pointer"
                                                >
                                                    {copiedStep === `template-${index}` ? (
                                                        <Check className="h-3 w-3 text-green-400" />
                                                    ) : (
                                                        <Copy className="h-3 w-3 text-slate-400" />
                                                    )}
                                                </button>
                                            </div>
                                        </div>
                                    ))}
                                </div>

                                {/* Or create from scratch */}
                                <div className="border-t border-slate-200 pt-6">
                                    <h4 className="font-semibold text-slate-900 mb-2">Or create from scratch:</h4>
                                    <div className="bg-slate-900 rounded-lg p-4 relative">
                                        <code className="text-green-400 text-sm font-mono">
                                            cohort-bridge init --interactive
                                        </code>
                                        <button
                                            onClick={() => copyToClipboard('cohort-bridge init --interactive', 'scratch')}
                                            className="absolute top-2 right-2 p-2 hover:bg-slate-700 rounded transition-colors cursor-pointer"
                                        >
                                            {copiedStep === 'scratch' ? (
                                                <Check className="h-4 w-4 text-green-400" />
                                            ) : (
                                                <Copy className="h-4 w-4 text-slate-400" />
                                            )}
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Step 4: Alternative - Use Web Builder */}
                    <div className="bg-gradient-to-r from-blue-600 to-indigo-600 rounded-xl shadow-lg p-8 text-white">
                        <div className="flex items-start space-x-4">
                            <div className="w-12 h-12 bg-white bg-opacity-20 rounded-full flex items-center justify-center flex-shrink-0">
                                <FileText className="h-6 w-6 text-white" />
                            </div>
                            <div className="flex-1">
                                <h3 className="text-xl font-semibold mb-3">
                                    Alternative: Use Our Web-Based Configuration Builder
                                </h3>
                                <p className="text-blue-100 mb-6">
                                    Prefer a visual interface? Use our web-based configuration builder to create
                                    and download YAML configurations without installing the CLI.
                                </p>
                                <button
                                    onClick={() => router.push('/config')}
                                    className="inline-flex items-center space-x-2 bg-white text-blue-600 px-6 py-3 rounded-lg hover:bg-blue-50 transition-colors cursor-pointer font-medium"
                                >
                                    <span>Try Configuration Builder</span>
                                    <ChevronRight className="h-4 w-4" />
                                </button>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Next Steps */}
                <div className="mt-16 text-center">
                    <h3 className="text-2xl font-bold text-slate-900 mb-4">What's Next?</h3>
                    <p className="text-slate-600 mb-6">
                        Once you have your configuration file, you can run CohortBridge and start linking records!
                    </p>
                    <div className="bg-slate-900 rounded-lg p-4 inline-block">
                        <code className="text-green-400 text-sm font-mono">
                            cohort-bridge run --config your-config.yaml
                        </code>
                    </div>
                </div>
            </div>
        </div>
    );
} 