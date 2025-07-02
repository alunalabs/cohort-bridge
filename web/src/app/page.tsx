'use client';

import { useRouter } from 'next/navigation';
import { Settings, Terminal, Github, Download, Shield, Database, FileText, Network } from 'lucide-react';

export default function HomePage() {
    const router = useRouter();

    const configTypes = [
        {
            id: 'basic',
            title: 'Basic Configuration',
            description: 'Simple setup for privacy-preserving record linkage with data normalization',
            icon: Settings,
            path: '/config/basic',
            features: ['CSV File Support', 'Data Normalization', 'Basic Privacy'],
            gradient: 'from-blue-500 to-purple-500'
        },
        {
            id: 'secure',
            title: 'Secure Configuration',
            description: 'Maximum security with comprehensive logging and audit trails',
            icon: Shield,
            path: '/config/secure',
            features: ['Enhanced Security', 'Audit Logging', 'IP Restrictions'],
            gradient: 'from-green-500 to-blue-500'
        },
        {
            id: 'postgres',
            title: 'PostgreSQL Configuration',
            description: 'Connect to PostgreSQL databases with enhanced security',
            icon: Database,
            path: '/config/postgres',
            features: ['Database Integration', 'Custom Schemas', 'Scalable Processing'],
            gradient: 'from-purple-500 to-pink-500'
        },
        {
            id: 'tokenized',
            title: 'Tokenized Configuration',
            description: 'Work with pre-tokenized data files for enhanced privacy',
            icon: FileText,
            path: '/config/tokenized',
            features: ['Pre-tokenized Data', 'Enhanced Privacy', 'Secure Processing'],
            gradient: 'from-teal-500 to-green-500'
        },
        {
            id: 'network',
            title: 'Network Configuration',
            description: 'Multi-party network setup with advanced matching algorithms',
            icon: Network,
            path: '/config/network',
            features: ['Multi-party Setup', 'Advanced Matching', 'Network Security'],
            gradient: 'from-orange-500 to-red-500'
        }
    ];

    return (
        <div className="min-h-screen bg-gray-50">
            {/* Header */}
            <header className="border-b border-gray-200 bg-white">
                <div className="max-w-6xl mx-auto px-4">
                    <div className="flex justify-between items-center py-4">
                        <div className="flex items-center space-x-3">
                            <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-500 rounded flex items-center justify-center">
                                <Settings className="h-4 w-4 text-white" />
                            </div>
                            <div>
                                <h1 className="text-lg font-bold text-gray-900">CohortBridge</h1>
                                <p className="text-xs text-gray-500">Privacy-Preserving Record Linkage</p>
                            </div>
                        </div>
                        <div className="flex items-center space-x-4">
                            <button
                                onClick={() => router.push('/docs')}
                                className="text-gray-600 hover:text-gray-900 transition-colors text-sm font-medium"
                            >
                                Documentation
                            </button>
                            <a
                                href="https://github.com/alunalabs/cohort-bridge"
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-gray-600 hover:text-gray-900 transition-colors text-sm font-medium"
                            >
                                GitHub
                            </a>
                        </div>
                    </div>
                </div>
            </header>

            {/* Hero Section */}
            <div className="max-w-6xl mx-auto px-4 py-16">
                <div className="text-center mb-16">
                    <h2 className="text-4xl font-bold text-gray-900 mb-4">
                        Privacy-Preserving Record Linkage
                    </h2>
                    <p className="text-xl text-gray-600 max-w-3xl mx-auto mb-8">
                        Securely identify matching patient records across healthcare institutions without sharing sensitive PHI data.
                        Configure your setup visually with our interactive configuration builder.
                    </p>
                    <div className="flex justify-center space-x-4">
                        <button
                            onClick={() => router.push('/config')}
                            className="inline-flex items-center px-6 py-3 bg-gradient-to-r from-blue-500 to-purple-500 text-white rounded-lg font-medium hover:from-blue-600 hover:to-purple-600 transition-all shadow-lg"
                        >
                            <Settings className="mr-2 h-5 w-5" />
                            Build Configuration
                        </button>
                        <button
                            onClick={() => router.push('/docs')}
                            className="inline-flex items-center px-6 py-3 border border-gray-300 text-gray-700 rounded-lg font-medium hover:border-gray-400 hover:text-gray-900 transition-colors bg-white"
                        >
                            <FileText className="mr-2 h-5 w-5" />
                            View Documentation
                        </button>
                        <button
                            onClick={() => router.push('/docs')}
                            className="inline-flex items-center px-6 py-3 border border-gray-300 text-gray-700 rounded-lg font-medium hover:border-gray-400 hover:text-gray-900 transition-colors bg-white"
                        >
                            <Terminal className="mr-2 h-5 w-5" />
                            Install CLI
                        </button>
                    </div>
                </div>

                {/* Configuration Types */}
                <div className="mb-16">
                    <div className="text-center mb-8">
                        <h3 className="text-2xl font-bold text-gray-900 mb-3">
                            Choose Your Configuration Type
                        </h3>
                        <p className="text-gray-600">
                            Select the configuration that best matches your deployment scenario
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {configTypes.map((config) => {
                            const IconComponent = config.icon;
                            return (
                                <div
                                    key={config.id}
                                    onClick={() => router.push(config.path)}
                                    className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 hover:shadow-lg transition-all cursor-pointer group"
                                >
                                    <div className="flex items-center space-x-3 mb-4">
                                        <div className={`w-10 h-10 bg-gradient-to-r ${config.gradient} rounded-lg flex items-center justify-center`}>
                                            <IconComponent className="h-5 w-5 text-white" />
                                        </div>
                                        <h4 className="text-lg font-semibold text-gray-900 group-hover:text-blue-600 transition-colors">
                                            {config.title}
                                        </h4>
                                    </div>
                                    <p className="text-gray-600 text-sm mb-4 leading-relaxed">
                                        {config.description}
                                    </p>
                                    <div className="space-y-2">
                                        {config.features.map((feature, index) => (
                                            <div key={index} className="flex items-center text-sm text-gray-500">
                                                <div className="w-1.5 h-1.5 bg-gray-400 rounded-full mr-2"></div>
                                                {feature}
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                </div>

                {/* Features Grid */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-16">
                    <div className="text-center">
                        <div className="w-16 h-16 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full flex items-center justify-center mx-auto mb-4">
                            <Settings className="h-8 w-8 text-white" />
                        </div>
                        <h4 className="text-lg font-semibold text-gray-900 mb-2">Visual Configuration</h4>
                        <p className="text-gray-600 text-sm leading-relaxed">
                            Build complex YAML configurations through intuitive forms. No manual editing required.
                        </p>
                    </div>

                    <div className="text-center">
                        <div className="w-16 h-16 bg-gradient-to-r from-green-500 to-blue-500 rounded-full flex items-center justify-center mx-auto mb-4">
                            <Shield className="h-8 w-8 text-white" />
                        </div>
                        <h4 className="text-lg font-semibold text-gray-900 mb-2">HIPAA Compliant</h4>
                        <p className="text-gray-600 text-sm leading-relaxed">
                            Built-in privacy protection with differential privacy, secure encryption, and audit logging.
                        </p>
                    </div>

                    <div className="text-center">
                        <div className="w-16 h-16 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full flex items-center justify-center mx-auto mb-4">
                            <Download className="h-8 w-8 text-white" />
                        </div>
                        <h4 className="text-lg font-semibold text-gray-900 mb-2">Production Ready</h4>
                        <p className="text-gray-600 text-sm leading-relaxed">
                            Export configurations that work directly with the CohortBridge CLI. No additional setup needed.
                        </p>
                    </div>
                </div>

                {/* Getting Started CTA */}
                <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-8 text-center">
                    <div className="w-16 h-16 bg-gradient-to-r from-gray-600 to-gray-700 rounded-full flex items-center justify-center mx-auto mb-4">
                        <Terminal className="h-8 w-8 text-white" />
                    </div>
                    <h3 className="text-xl font-bold text-gray-900 mb-2">Ready to Get Started?</h3>
                    <p className="text-gray-600 max-w-2xl mx-auto mb-6 leading-relaxed">
                        Install the CLI tool locally or use our web interface to generate your configuration files.
                        Complete setup in minutes.
                    </p>
                    <div className="flex justify-center space-x-3">
                        <button
                            onClick={() => router.push('/config')}
                            className="inline-flex items-center px-6 py-3 bg-gradient-to-r from-blue-500 to-purple-500 text-white rounded-lg font-medium hover:from-blue-600 hover:to-purple-600 transition-all"
                        >
                            Start Configuring
                        </button>
                        <button
                            onClick={() => router.push('/docs')}
                            className="inline-flex items-center px-6 py-3 bg-gray-900 text-white rounded-lg font-medium hover:bg-gray-800 transition-colors"
                        >
                            <Github className="mr-2 h-4 w-4" />
                            Install CLI
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
} 