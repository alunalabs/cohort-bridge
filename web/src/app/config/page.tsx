'use client';

import { useRouter } from 'next/navigation';
import { ArrowLeft, Settings, Shield, Download, Terminal } from 'lucide-react';
import ConfigurationSelector from '../../components/ConfigurationSelector';

export default function ConfigPage() {
    const router = useRouter();

    return (
        <div className="min-h-screen bg-white">
            {/* Header */}
            <header className="border-b border-gray-200 bg-white">
                <div className="max-w-6xl mx-auto px-6">
                    <div className="flex justify-between items-center py-6">
                        <div className="flex items-center space-x-4">
                            <button
                                onClick={() => router.push('/')}
                                className="flex items-center space-x-2 text-gray-600 hover:text-gray-900 transition-colors"
                            >
                                <ArrowLeft className="h-5 w-5" />
                                <span className="font-medium">Back</span>
                            </button>
                        </div>
                        <div className="flex items-center space-x-3">
                            <div className="w-10 h-10 bg-gradient-to-r from-blue-600 to-purple-600 rounded-2xl flex items-center justify-center">
                                <Settings className="h-5 w-5 text-white" />
                            </div>
                            <div>
                                <h1 className="text-lg font-bold text-gray-900">CohortBridge</h1>
                                <p className="text-sm text-gray-600">Configuration Builder</p>
                            </div>
                        </div>
                    </div>
                </div>
            </header>

            {/* Content */}
            <div className="max-w-6xl mx-auto px-6 py-20">
                {/* Configuration Selector */}
                <ConfigurationSelector />

                {/* Features Section */}
                <div className="mt-32 bg-gray-50 rounded-3xl p-12 lg:p-16">
                    <div className="text-center mb-16">
                        <h3 className="text-3xl lg:text-4xl font-bold text-gray-900 mb-4">
                            Why use our builder?
                        </h3>
                        <p className="text-xl text-gray-600">
                            Built for developers, designed for simplicity
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-12">
                        <div className="text-center group">
                            <div className="w-20 h-20 bg-blue-100 rounded-3xl flex items-center justify-center mx-auto mb-6 group-hover:bg-blue-600 transition-colors duration-300">
                                <Settings className="h-10 w-10 text-blue-600 group-hover:text-white transition-colors duration-300" />
                            </div>
                            <h4 className="text-xl font-bold text-gray-900 mb-4">Visual Interface</h4>
                            <p className="text-gray-600 leading-relaxed">
                                No more manual YAML editing. Build configurations with an intuitive form interface.
                            </p>
                        </div>

                        <div className="text-center group">
                            <div className="w-20 h-20 bg-green-100 rounded-3xl flex items-center justify-center mx-auto mb-6 group-hover:bg-green-600 transition-colors duration-300">
                                <Shield className="h-10 w-10 text-green-600 group-hover:text-white transition-colors duration-300" />
                            </div>
                            <h4 className="text-xl font-bold text-gray-900 mb-4">Validation</h4>
                            <p className="text-gray-600 leading-relaxed">
                                Built-in validation ensures your configurations are correct before export.
                            </p>
                        </div>

                        <div className="text-center group">
                            <div className="w-20 h-20 bg-purple-100 rounded-3xl flex items-center justify-center mx-auto mb-6 group-hover:bg-purple-600 transition-colors duration-300">
                                <Download className="h-10 w-10 text-purple-600 group-hover:text-white transition-colors duration-300" />
                            </div>
                            <h4 className="text-xl font-bold text-gray-900 mb-4">Export Ready</h4>
                            <p className="text-gray-600 leading-relaxed">
                                Download production-ready YAML files that work seamlessly with CohortBridge.
                            </p>
                        </div>
                    </div>
                </div>

                {/* CLI Alternative Section */}
                <div className="mt-20 bg-gradient-to-r from-gray-900 to-gray-800 rounded-3xl p-12 lg:p-16 text-white">
                    <div className="text-center">
                        <div className="w-20 h-20 bg-white bg-opacity-20 rounded-3xl flex items-center justify-center mx-auto mb-8">
                            <Terminal className="h-10 w-10 text-white" />
                        </div>
                        <h3 className="text-3xl lg:text-4xl font-bold mb-6">Prefer command line?</h3>
                        <p className="text-xl text-gray-300 max-w-2xl mx-auto mb-8 leading-relaxed">
                            For developers who prefer working with CLI tools, you can install CohortBridge locally
                            and create configurations using our command-line interface.
                        </p>
                        <button
                            onClick={() => router.push('/get-started')}
                            className="inline-flex items-center px-8 py-4 bg-white text-gray-900 rounded-2xl hover:bg-gray-100 transition-colors font-semibold text-lg shadow-lg"
                        >
                            CLI Installation Guide
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
} 