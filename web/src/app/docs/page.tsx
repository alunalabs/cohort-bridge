'use client';

import { useRouter } from 'next/navigation';
import {
    Settings,
    Github,
    BookOpen,
    Zap,
    Lock,
    Search,
    Network
} from 'lucide-react';
import { useState } from 'react';
import OverviewTab from '@/components/docs/OverviewTab';
import ConfigurationTab from '@/components/docs/ConfigurationTab';
import PPRLTab from '@/components/docs/PPRLTab';
import TokenizeTab from '@/components/docs/TokenizeTab';
import ValidateTab from '@/components/docs/ValidateTab';
import ExamplesTab from '@/components/docs/ExamplesTab';

export default function DocsPage() {
    const router = useRouter();
    const [activeTab, setActiveTab] = useState('configuration');

    const tabs = [
        { id: 'overview', name: 'Overview', icon: BookOpen },
        { id: 'configuration', name: 'Configuration', icon: Settings },
        { id: 'pprl', name: 'Peer-to-Peer Matching', icon: Network },
        { id: 'tokenize', name: 'Data Tokenization', icon: Lock },
        { id: 'validate', name: 'Result Validation', icon: Search },
        { id: 'examples', name: 'Examples', icon: Zap },
    ];

    const renderTabContent = () => {
        switch (activeTab) {
            case 'overview':
                return <OverviewTab />;
            case 'configuration':
                return <ConfigurationTab />;
            case 'pprl':
                return <PPRLTab />;
            case 'tokenize':
                return <TokenizeTab />;
            case 'validate':
                return <ValidateTab />;
            case 'examples':
                return <ExamplesTab />;
            default:
                return null;
        }
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-50 via-white to-slate-100/50">
            {/* Header */}
            <header className="border-b border-slate-200/60 bg-white/80 backdrop-blur-md sticky top-0 z-50">
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
                                onClick={() => router.push('/')}
                                className="text-gray-600 hover:text-gray-900 transition-colors text-sm font-medium"
                            >
                                Home
                            </button>
                            <button
                                onClick={() => router.push('/config')}
                                className="text-gray-600 hover:text-gray-900 transition-colors text-sm font-medium"
                            >
                                Config Builder
                            </button>
                            <a
                                href="https://github.com/alunalabs/cohort-bridge"
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-gray-600 hover:text-gray-900 transition-colors"
                            >
                                <Github className="h-5 w-5" />
                            </a>
                        </div>
                    </div>
                </div>
            </header>

            {/* Hero Section */}
            <div className="bg-gradient-to-r from-blue-500 to-purple-500 text-white">
                <div className="max-w-6xl mx-auto px-4 py-16">
                    <div className="text-center">
                        <h2 className="text-4xl font-bold mb-4">
                            CohortBridge Documentation
                        </h2>
                        <p className="text-xl opacity-90 max-w-3xl mx-auto">
                            Complete guide to using the CohortBridge CLI tool for privacy-preserving record linkage.
                            Learn how to tokenize data, run peer-to-peer matching, and validate results.
                        </p>
                    </div>
                </div>
            </div>

            {/* Main Content */}
            <div className="max-w-6xl mx-auto px-4 py-12">
                {/* Tab Navigation */}
                <div className="mb-8">
                    <div className="border-b border-gray-200">
                        <nav className="-mb-px flex space-x-8 overflow-x-auto">
                            {tabs.map((tab) => {
                                const Icon = tab.icon;
                                return (
                                    <button
                                        key={tab.id}
                                        onClick={() => setActiveTab(tab.id)}
                                        className={`flex items-center space-x-2 py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap transition-colors ${activeTab === tab.id
                                            ? 'border-blue-500 text-blue-600'
                                            : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                                            }`}
                                    >
                                        <Icon className="h-4 w-4" />
                                        <span>{tab.name}</span>
                                    </button>
                                );
                            })}
                        </nav>
                    </div>
                </div>

                {/* Tab Content */}
                <div className="bg-white/70 backdrop-blur-sm rounded-xl shadow-lg border border-gray-200 p-8">
                    {renderTabContent()}
                </div>

                {/* Footer */}
                <div className="mt-16 pt-8 border-t border-gray-200">
                    <div className="text-center">
                        <h3 className="text-lg font-semibold text-gray-900 mb-4">Need Help?</h3>
                        <div className="flex justify-center space-x-6 text-sm">
                            <a href="https://github.com/alunalabs/cohort-bridge/issues" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:text-blue-800">
                                Report Issues
                            </a>
                            <a href="https://github.com/alunalabs/cohort-bridge/discussions" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:text-blue-800">
                                Discussions
                            </a>
                            <button onClick={() => router.push('/config')} className="text-blue-600 hover:text-blue-800">
                                Config Builder
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
} 