'use client';

import { useRouter } from 'next/navigation';
import { Settings, Terminal, Github, Download, Shield, Database, FileText, Network, ArrowRight } from 'lucide-react';

export default function HomePage() {
    const router = useRouter();

    const configTypes = [
        {
            id: 'basic',
            title: 'Basic Configuration',
            description: 'Quick start for standard CSV-based record matching',
            icon: Settings,
            path: '/config/basic',
            features: ['CSV File Support', 'Standard Matching', 'Simple Setup'],
            gradient: 'from-blue-500 to-purple-500',
            useCase: 'Best for first-time users or simple CSV file matching scenarios'
        },
        {
            id: 'postgres',
            title: 'PostgreSQL Template',
            description: 'Direct database connections for enterprise environments',
            icon: Database,
            path: '/config/postgres',
            features: ['Database Integration', 'Enterprise Scale', 'SQL Queries'],
            gradient: 'from-purple-500 to-pink-500',
            useCase: 'When your data lives in PostgreSQL and you need production-scale processing'
        },
        {
            id: 'tokenized',
            title: 'Pre-tokenized Template',
            description: 'Work with already processed and anonymized datasets',
            icon: Terminal,
            path: '/config/tokenized',
            features: ['Skip Tokenization', 'Fast Processing', 'Pre-anonymized'],
            gradient: 'from-teal-500 to-green-500',
            useCase: 'When you already have tokenized data or need to skip preprocessing steps'
        },
    ];

    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-white relative overflow-hidden">
            {/* Tech background patterns */}
            <div className="absolute inset-0">
                {/* Grid pattern */}
                <div className="absolute top-0 left-0 w-full h-full bg-[linear-gradient(to_right,#3b82f610_1px,transparent_1px),linear-gradient(to_bottom,#3b82f610_1px,transparent_1px)] bg-[size:32px_32px]"></div>
                {/* Circuit-like pattern */}
                <div className="absolute top-0 left-0 w-full h-full opacity-20">
                    <div className="absolute top-10 left-10 w-32 h-32 border border-blue-500/20 rounded-lg"></div>
                    <div className="absolute top-20 right-20 w-24 h-24 border border-blue-600/20 rounded-full"></div>
                    <div className="absolute bottom-20 left-20 w-40 h-40 border border-blue-400/20 rounded-xl"></div>
                    <div className="absolute bottom-32 right-32 w-28 h-28 border border-blue-700/20 rounded-lg"></div>
                </div>
                {/* Glow effects */}
                <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/5 rounded-full blur-3xl"></div>
                <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-blue-600/5 rounded-full blur-3xl"></div>
            </div>

            {/* Header */}
            <header className="border-b border-blue-200/50 bg-white/80 backdrop-blur-sm relative z-10">
                <div className="max-w-6xl mx-auto px-4">
                    <div className="flex justify-between items-center py-4">
                        <div className="flex items-center space-x-3">
                            <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-blue-600 rounded flex items-center justify-center shadow-lg">
                                <Settings className="h-4 w-4 text-white" />
                            </div>
                            <div>
                                <h1 className="text-lg font-bold text-gray-900">CohortBridge</h1>
                                <p className="text-xs text-blue-600">Privacy-Preserving Record Linkage</p>
                            </div>
                        </div>
                        <div className="flex items-center space-x-4">
                            <button
                                onClick={() => router.push('/docs')}
                                className="text-gray-600 hover:text-blue-600 transition-colors text-sm font-medium"
                            >
                                Documentation
                            </button>
                            <button
                                onClick={() => router.push('/config/basic')}
                                className="text-gray-600 hover:text-blue-600 transition-colors text-sm font-medium"
                            >
                                Configuration
                            </button>
                            <a
                                href="https://github.com/alunalabs/cohort-bridge"
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-gray-600 hover:text-blue-600 transition-colors text-sm font-medium"
                            >
                                GitHub
                            </a>
                        </div>
                    </div>
                </div>
            </header>

            {/* Hero Section */}
            <div className="max-w-6xl mx-auto px-4 py-16 relative z-10">
                <div className="text-center mb-16">
                    <h2 className="text-5xl font-bold text-gray-900 mb-6 tracking-tight">
                        Direct Peer-to-Peer Record Linkage
                    </h2>
                    <p className="text-xl text-gray-600 max-w-3xl mx-auto mb-8 leading-relaxed">
                        Cut out the middleman. Connect directly with your partners to identify matching records
                        without exposing sensitive data to third parties. Share only what you need with who you trust.
                    </p>
                    <div className="flex justify-center space-x-4">
                        <button
                            onClick={() => router.push('/docs')}
                            className="inline-flex items-center px-8 py-4 bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-lg font-medium hover:from-blue-600 hover:to-blue-700 transition-all shadow-lg shadow-blue-500/25"
                        >
                            <FileText className="mr-2 h-5 w-5" />
                            Get Started
                        </button>
                        <button
                            onClick={() => router.push('/docs')}
                            className="inline-flex items-center px-6 py-3 border border-blue-200 text-gray-700 rounded-lg font-medium hover:border-blue-300 hover:bg-blue-50 transition-colors backdrop-blur-sm"
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
                            Configuration Starter Templates
                        </h3>
                        <p className="text-gray-600 mb-2">
                            All templates provide complete flexibility - some just provide better starting points for specific scenarios
                        </p>
                        <p className="text-sm text-gray-500">
                            Every template can be customized to do anything with the visual configuration builder
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {configTypes.map((config) => {
                            const IconComponent = config.icon;
                            return (
                                <div
                                    key={config.id}
                                    onClick={() => router.push(config.path)}
                                    className="bg-white/80 backdrop-blur-sm rounded-xl shadow-lg border border-blue-200/50 p-6 hover:shadow-xl hover:bg-white/90 hover:border-blue-300/50 transition-all cursor-pointer group relative overflow-hidden"
                                >
                                    {/* Tech overlay */}
                                    <div className="absolute inset-0 bg-gradient-to-br from-blue-500/5 to-blue-600/5 opacity-0 group-hover:opacity-100 transition-opacity"></div>

                                    <div className="relative z-10">
                                        <div className="flex items-center space-x-3 mb-4">
                                            <div className={`w-10 h-10 bg-gradient-to-r ${config.gradient} rounded-lg flex items-center justify-center shadow-lg`}>
                                                <IconComponent className="h-5 w-5 text-white" />
                                            </div>
                                            <h4 className="text-lg font-semibold text-gray-900 group-hover:text-blue-600 transition-colors">
                                                {config.title}
                                            </h4>
                                        </div>
                                        <p className="text-gray-600 text-sm mb-3 leading-relaxed">
                                            {config.description}
                                        </p>

                                        {/* Use case */}
                                        <div className="bg-blue-50/80 rounded-lg p-3 mb-4 border border-blue-100">
                                            <p className="text-xs text-blue-700 font-medium mb-1">Optimized for:</p>
                                            <p className="text-xs text-blue-600 leading-relaxed">{config.useCase}</p>
                                        </div>

                                        <div className="text-xs text-gray-500 italic">
                                            Fully customizable with the visual configuration builder
                                        </div>
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                </div>

                {/* PPRL Connection Process */}
                <div className="mb-16">
                    <div className="bg-white/80 backdrop-blur-sm rounded-xl shadow-lg border border-blue-200/50 p-8 relative overflow-hidden">
                        {/* Tech background */}
                        <div className="absolute inset-0 opacity-10">
                            <div className="absolute top-0 left-0 w-full h-full bg-[radial-gradient(circle_at_50%_50%,theme(colors.blue.500),transparent_50%)]"></div>
                            <div className="absolute top-4 left-4 w-16 h-16 border border-blue-400/30 rounded-lg animate-pulse"></div>
                            <div className="absolute bottom-4 right-4 w-20 h-20 border border-blue-500/30 rounded-full animate-pulse"></div>
                        </div>

                        <div className="relative z-10">
                            <div className="text-center mb-8">
                                <h3 className="text-2xl font-bold text-gray-900 mb-3">
                                    How PPRL Connection Works
                                </h3>
                                <p className="text-gray-600 max-w-3xl mx-auto">
                                    Direct peer-to-peer connection ensures your data never touches a third party
                                </p>
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-7 gap-4 items-center">
                                {/* Step 1 */}
                                <div className="text-center">
                                    <div className="w-12 h-12 bg-gradient-to-r from-blue-500 to-blue-600 rounded-full flex items-center justify-center mx-auto mb-3 shadow-lg">
                                        <span className="text-white font-bold">1</span>
                                    </div>
                                    <p className="text-sm font-medium text-gray-900 mb-1">Tokenize</p>
                                    <p className="text-xs text-gray-600">Process your local data</p>
                                </div>

                                {/* Arrow */}
                                <div className="hidden md:flex justify-center">
                                    <ArrowRight className="h-4 w-4 text-teal-500" />
                                </div>

                                {/* Step 2 */}
                                <div className="text-center">
                                    <div className="w-12 h-12 bg-gradient-to-r from-teal-500 to-teal-600 rounded-full flex items-center justify-center mx-auto mb-3 shadow-lg">
                                        <span className="text-white font-bold">2</span>
                                    </div>
                                    <p className="text-sm font-medium text-gray-900 mb-1">Connect</p>
                                    <p className="text-xs text-gray-600">Direct peer connection</p>
                                </div>

                                {/* Arrow */}
                                <div className="hidden md:flex justify-center">
                                    <ArrowRight className="h-4 w-4 text-purple-500" />
                                </div>

                                {/* Step 3 */}
                                <div className="text-center">
                                    <div className="w-12 h-12 bg-gradient-to-r from-purple-500 to-purple-600 rounded-full flex items-center justify-center mx-auto mb-3 shadow-lg">
                                        <span className="text-white font-bold">3</span>
                                    </div>
                                    <p className="text-sm font-medium text-gray-900 mb-1">Exchange</p>
                                    <p className="text-xs text-gray-600">Secure token exchange</p>
                                </div>

                                {/* Arrow */}
                                <div className="hidden md:flex justify-center">
                                    <ArrowRight className="h-4 w-4 text-green-500" />
                                </div>

                                {/* Step 4 */}
                                <div className="text-center">
                                    <div className="w-12 h-12 bg-gradient-to-r from-green-500 to-green-600 rounded-full flex items-center justify-center mx-auto mb-3 shadow-lg">
                                        <span className="text-white font-bold">4</span>
                                    </div>
                                    <p className="text-sm font-medium text-gray-900 mb-1">Match</p>
                                    <p className="text-xs text-gray-600">Zero-knowledge matching</p>
                                </div>
                            </div>

                            <div className="mt-8 text-center">
                                <div className="inline-flex items-center space-x-2 bg-blue-50/80 rounded-lg px-4 py-2 border border-blue-200">
                                    <Shield className="h-4 w-4 text-blue-600" />
                                    <span className="text-sm text-blue-700 font-medium">No middleman • No data leakage • Direct control</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Features Grid */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-16 relative z-10">
                    <div className="text-center bg-white/80 backdrop-blur-sm rounded-xl p-6 border border-blue-200/50 hover:border-blue-300/50 transition-all group">
                        <div className="w-16 h-16 bg-gradient-to-r from-blue-500 to-blue-600 rounded-full flex items-center justify-center mx-auto mb-4 shadow-lg group-hover:shadow-blue-500/25 transition-all">
                            <Settings className="h-8 w-8 text-white" />
                        </div>
                        <h4 className="text-lg font-semibold text-gray-900 mb-2">Zero-Code Configuration</h4>
                        <p className="text-gray-600 text-sm leading-relaxed">
                            Build enterprise-grade YAML configurations through visual forms. No technical expertise required.
                        </p>
                    </div>

                    <div className="text-center bg-white/80 backdrop-blur-sm rounded-xl p-6 border border-green-200/50 hover:border-green-300/50 transition-all group">
                        <div className="w-16 h-16 bg-gradient-to-r from-green-500 to-green-600 rounded-full flex items-center justify-center mx-auto mb-4 shadow-lg group-hover:shadow-green-500/25 transition-all">
                            <Shield className="h-8 w-8 text-white" />
                        </div>
                        <h4 className="text-lg font-semibold text-gray-900 mb-2">Zero-Knowledge Protocols</h4>
                        <p className="text-gray-600 text-sm leading-relaxed">
                            Advanced cryptographic protocols ensure absolutely no data leakage beyond the final match results.
                        </p>
                    </div>

                    <div className="text-center bg-white/80 backdrop-blur-sm rounded-xl p-6 border border-purple-200/50 hover:border-purple-300/50 transition-all group">
                        <div className="w-16 h-16 bg-gradient-to-r from-purple-500 to-purple-600 rounded-full flex items-center justify-center mx-auto mb-4 shadow-lg group-hover:shadow-purple-500/25 transition-all">
                            <Network className="h-8 w-8 text-white" />
                        </div>
                        <h4 className="text-lg font-semibold text-gray-900 mb-2">Direct P2P Connection</h4>
                        <p className="text-gray-600 text-sm leading-relaxed">
                            Connect directly with your partners. No third-party servers, no data intermediaries, total control.
                        </p>
                    </div>
                </div>

                {/* Performance Metrics */}
                <div className="mb-16 relative z-10">
                    <div className="bg-white/80 backdrop-blur-sm rounded-xl shadow-lg border border-blue-200/50 p-8 relative overflow-hidden">
                        {/* Performance background */}
                        <div className="absolute inset-0 opacity-10">
                            <div className="absolute top-0 left-0 w-full h-full bg-[conic-gradient(from_0deg_at_50%_50%,theme(colors.green.500),theme(colors.blue.500),theme(colors.purple.500),theme(colors.green.500))]"></div>
                        </div>

                        <div className="relative z-10">
                            <div className="text-center mb-8">
                                <h3 className="text-2xl font-bold text-gray-900 mb-3">
                                    Proven Performance
                                </h3>
                                <p className="text-gray-600 max-w-2xl mx-auto">
                                    Industry-leading accuracy metrics validated across multiple datasets
                                </p>
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                                <div className="text-center">
                                    <div className="bg-green-50/80 rounded-lg p-6 border border-green-200">
                                        <div className="flex items-center justify-center mb-2">
                                            <span className="text-3xl font-mono font-bold text-green-600">0.999</span>
                                        </div>
                                        <p className="text-sm font-medium text-gray-900 mb-1">F1-Score</p>
                                        <p className="text-xs text-gray-600">Harmonic mean of precision and recall</p>
                                    </div>
                                </div>

                                <div className="text-center">
                                    <div className="bg-blue-50/80 rounded-lg p-6 border border-blue-200">
                                        <div className="flex items-center justify-center mb-2">
                                            <span className="text-3xl font-mono font-bold text-blue-600">0.999</span>
                                        </div>
                                        <p className="text-sm font-medium text-gray-900 mb-1">Precision</p>
                                        <p className="text-xs text-gray-600">True positives / all positive predictions</p>
                                    </div>
                                </div>

                                <div className="text-center">
                                    <div className="bg-purple-50/80 rounded-lg p-6 border border-purple-200">
                                        <div className="flex items-center justify-center mb-2">
                                            <span className="text-3xl font-mono font-bold text-purple-600">0.998</span>
                                        </div>
                                        <p className="text-sm font-medium text-gray-900 mb-1">Recall</p>
                                        <p className="text-xs text-gray-600">True positives / all actual positives</p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Getting Started CTA */}
                <div className="bg-white/80 backdrop-blur-sm rounded-xl shadow-lg border border-blue-200/50 p-8 text-center relative z-10 overflow-hidden">
                    {/* Tech background */}
                    <div className="absolute inset-0 opacity-10">
                        <div className="absolute top-0 left-0 w-full h-full bg-[conic-gradient(from_0deg_at_50%_50%,theme(colors.blue.500),theme(colors.blue.600),theme(colors.blue.700),theme(colors.blue.500))]"></div>
                        <div className="absolute top-8 left-8 w-12 h-12 border border-blue-400/30 rounded-lg animate-pulse"></div>
                        <div className="absolute bottom-8 right-8 w-16 h-16 border border-blue-500/30 rounded-full animate-pulse"></div>
                    </div>

                    <div className="relative z-10">
                        <div className="w-16 h-16 bg-gradient-to-r from-blue-500 to-blue-600 rounded-full flex items-center justify-center mx-auto mb-4 shadow-lg">
                            <Terminal className="h-8 w-8 text-white" />
                        </div>
                        <h3 className="text-xl font-bold text-gray-900 mb-2">Ready to Connect Directly?</h3>
                        <p className="text-gray-600 max-w-2xl mx-auto mb-6 leading-relaxed">
                            Skip the middleman and start matching records directly with your partners using zero-knowledge protocols.
                        </p>
                        <div className="flex justify-center space-x-3">
                            <button
                                onClick={() => router.push('/config/basic')}
                                className="inline-flex items-center px-6 py-3 bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-lg font-medium hover:from-blue-600 hover:to-blue-700 transition-all shadow-lg shadow-blue-500/25"
                            >
                                Build Configuration
                            </button>
                            <button
                                onClick={() => router.push('/docs')}
                                className="inline-flex items-center px-6 py-3 bg-white text-gray-700 rounded-lg font-medium hover:bg-gray-50 transition-colors shadow-lg border border-blue-200"
                            >
                                <Github className="mr-2 h-4 w-4" />
                                Install CLI
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
} 