'use client';

import { useRouter } from 'next/navigation';
import { ArrowLeft, Settings, Shield, Download, Terminal, Github } from 'lucide-react';
import ConfigurationSelector from '../../components/ConfigurationSelector';

export default function ConfigPage() {
    const router = useRouter();

    return (
        <div className="min-h-screen bg-gray-50">
            {/* Header */}
            <header className="border-b border-gray-200 bg-white">
                <div className="max-w-6xl mx-auto px-4">
                    <div className="flex justify-between items-center py-4">
                        <button
                            onClick={() => router.push('/')}
                            className="flex items-center space-x-2 text-gray-600 hover:text-gray-900 transition-colors"
                        >
                            <ArrowLeft className="h-4 w-4" />
                            <span className="text-sm font-medium">Back</span>
                        </button>
                        <div className="flex items-center space-x-2">
                            <div className="w-6 h-6 bg-gradient-to-r from-blue-500 to-purple-500 rounded flex items-center justify-center">
                                <Settings className="h-3 w-3 text-white" />
                            </div>
                            <div>
                                <h1 className="text-sm font-medium text-gray-900">Configuration Builder</h1>
                                <p className="text-xs text-gray-500">Visual YAML Generator</p>
                            </div>
                        </div>
                        <a
                            href="https://github.com/alunalabs/cohort-bridge#readme"
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-gray-600 hover:text-gray-900 transition-colors text-sm"
                        >
                            Documentation
                        </a>
                    </div>
                </div>
            </header>

            {/* Content */}
            <div className="max-w-6xl mx-auto px-4 py-12">
                {/* Configuration Selector */}
                <ConfigurationSelector />

                {/* Features Section */}
                <div className="mt-16 gradient-card rounded p-8 border">
                    <div className="text-center mb-8">
                        <h3 className="text-xl font-bold text-gray-900 mb-2">
                            Visual Configuration Builder
                        </h3>
                        <p className="text-gray-600 text-sm">
                            Form-based interface for complex YAML generation
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                        <div className="text-center">
                            <div className="w-12 h-12 bg-gradient-to-r from-blue-500 to-purple-500 rounded flex items-center justify-center mx-auto mb-3">
                                <Settings className="h-6 w-6 text-white" />
                            </div>
                            <h4 className="text-lg font-semibold text-gray-900 mb-2">Form Interface</h4>
                            <p className="text-gray-600 text-sm leading-relaxed">
                                No manual YAML editing. Build configurations through structured forms with validation.
                            </p>
                        </div>

                        <div className="text-center">
                            <div className="w-12 h-12 bg-gradient-to-r from-teal-500 to-green-500 rounded flex items-center justify-center mx-auto mb-3">
                                <Shield className="h-6 w-6 text-white" />
                            </div>
                            <h4 className="text-lg font-semibold text-gray-900 mb-2">Schema Validation</h4>
                            <p className="text-gray-600 text-sm leading-relaxed">
                                Real-time validation ensures configuration correctness before export.
                            </p>
                        </div>

                        <div className="text-center">
                            <div className="w-12 h-12 bg-gradient-to-r from-purple-500 to-pink-500 rounded flex items-center justify-center mx-auto mb-3">
                                <Download className="h-6 w-6 text-white" />
                            </div>
                            <h4 className="text-lg font-semibold text-gray-900 mb-2">Production Ready</h4>
                            <p className="text-gray-600 text-sm leading-relaxed">
                                Export YAML files that work directly with CohortBridge CLI.
                            </p>
                        </div>
                    </div>
                </div>

                {/* CLI Alternative */}
                <div className="mt-12 gradient-card rounded p-8 text-center border">
                    <div className="w-12 h-12 bg-gradient-to-r from-gray-600 to-gray-700 rounded flex items-center justify-center mx-auto mb-4">
                        <Terminal className="h-6 w-6 text-white" />
                    </div>
                    <h3 className="text-xl font-bold text-gray-900 mb-2">CLI Alternative</h3>
                    <p className="text-gray-600 text-sm max-w-2xl mx-auto mb-6 leading-relaxed">
                        For developers who prefer command-line tools, install CohortBridge locally
                        and generate configurations using built-in templates.
                    </p>
                    <div className="flex justify-center space-x-3">
                        <button
                            onClick={() => router.push('/get-started')}
                            className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-blue-500 to-purple-500 text-white rounded text-sm font-medium hover:from-blue-600 hover:to-purple-600 transition-all"
                        >
                            Install CLI
                        </button>
                        <a
                            href="https://github.com/alunalabs/cohort-bridge"
                            target="_blank"
                            rel="noopener noreferrer"
                            className="inline-flex items-center px-4 py-2 bg-gray-900 text-white rounded text-sm font-medium hover:bg-gray-800 transition-colors"
                        >
                            <Github className="mr-1.5 h-4 w-4" />
                            View Source
                        </a>
                    </div>
                </div>
            </div>
        </div>
    );
} 