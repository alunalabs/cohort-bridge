'use client';

import { useState, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { Settings, Database, Shield, Network, Users, Upload, ArrowRight } from 'lucide-react';
import yaml from 'js-yaml';

interface ConfigurationSelectorProps {
    title?: string;
    description?: string;
    showTitle?: boolean;
}

export default function ConfigurationSelector({
    title = "Choose Your Configuration",
    description = "Select the configuration type that best fits your use case and requirements, or import an existing configuration file to edit.",
    showTitle = true
}: ConfigurationSelectorProps) {
    const router = useRouter();
    const fileInputRef = useRef<HTMLInputElement>(null);
    const [selectedConfig, setSelectedConfig] = useState<string | null>(null);

    const configTypes = [
        {
            id: 'basic',
            name: 'Basic Configuration',
            description: 'Essential settings for peer-to-peer patient matching',
            icon: Settings,
            color: 'bg-blue-500',
            hoverColor: 'hover:bg-blue-100',
            path: '/config/basic'
        },
        {
            id: 'postgres',
            name: 'PostgreSQL Configuration',
            description: 'Enterprise database configuration with PostgreSQL',
            icon: Database,
            color: 'bg-green-500',
            hoverColor: 'hover:bg-green-100',
            path: '/config/postgres'
        },
        {
            id: 'secure',
            name: 'Secure Configuration',
            description: 'Enhanced security with access controls and monitoring',
            icon: Shield,
            color: 'bg-red-500',
            hoverColor: 'hover:bg-red-100',
            path: '/config/secure'
        },
        {
            id: 'tokenized',
            name: 'Tokenized Configuration',
            description: 'Advanced privacy with tokenized patient data',
            icon: Users,
            color: 'bg-purple-500',
            hoverColor: 'hover:bg-purple-100',
            path: '/config/tokenized'
        },
        {
            id: 'network',
            name: 'Network Configuration',
            description: 'Advanced networking and performance tuning',
            icon: Network,
            color: 'bg-orange-500',
            hoverColor: 'hover:bg-orange-100',
            path: '/config/network'
        }
    ];

    const handleConfigSelect = (config: typeof configTypes[0]) => {
        setSelectedConfig(config.id);
        router.push(config.path);
    };

    const handleImportConfig = () => {
        fileInputRef.current?.click();
    };

    const handleFileImport = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (!file) return;

        try {
            const text = await file.text();
            const parsedConfig = yaml.load(text) as Record<string, any>;

            if (!parsedConfig || typeof parsedConfig !== 'object') {
                alert('Invalid YAML file format. Please check your file and try again.');
                return;
            }

            // Determine the best configuration page based on the imported data
            let targetPage = '/config/basic'; // default

            if (parsedConfig.security || parsedConfig.timeouts || parsedConfig.logging) {
                targetPage = '/config/secure';
            } else if (parsedConfig.database?.type === 'postgres') {
                targetPage = '/config/postgres';
            } else if (parsedConfig.database?.is_tokenized) {
                targetPage = '/config/tokenized';
            } else if (parsedConfig.matching || parsedConfig.peer?.advanced_settings) {
                targetPage = '/config/network';
            }

            // Store the imported config in sessionStorage so the target page can pick it up
            sessionStorage.setItem('importedConfig', JSON.stringify(parsedConfig));
            sessionStorage.setItem('isImporting', 'true');

            // Navigate to the appropriate configuration page
            router.push(targetPage);

        } catch (error) {
            console.error('Error importing config:', error);
            alert('Failed to parse YAML file. Please check the file format and try again.');
        }

        // Reset file input
        if (fileInputRef.current) {
            fileInputRef.current.value = '';
        }
    };

    return (
        <div className="w-full">
            {/* Hidden file input */}
            <input
                ref={fileInputRef}
                type="file"
                accept=".yaml,.yml"
                onChange={handleFileImport}
                style={{ display: 'none' }}
            />

            {showTitle && (
                <div className="text-center mb-12">
                    <h2 className="text-3xl lg:text-4xl font-bold text-gray-900 mb-4">
                        {title}
                    </h2>
                    <p className="text-xl text-gray-600 max-w-3xl mx-auto leading-relaxed">
                        {description}
                    </p>
                </div>
            )}

            {/* Import Option - Prominent */}
            <div className="mb-10">
                <div
                    onClick={handleImportConfig}
                    className="group relative bg-gradient-to-r from-orange-50 to-amber-50 border-2 border-orange-200 rounded-3xl p-8 hover:border-orange-300 transition-all duration-300 cursor-pointer hover:shadow-lg"
                >
                    <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-6">
                            <div className="w-16 h-16 bg-orange-500 rounded-2xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                                <Upload className="h-8 w-8 text-white" />
                            </div>
                            <div>
                                <h3 className="text-2xl font-bold text-gray-900 mb-2">
                                    Import Configuration
                                </h3>
                                <p className="text-gray-600 text-lg">
                                    Upload an existing YAML file to customize
                                </p>
                            </div>
                        </div>
                        <div className="flex items-center text-orange-600 font-semibold text-lg">
                            <span>Browse files</span>
                            <ArrowRight className="h-5 w-5 ml-2 group-hover:translate-x-1 transition-transform" />
                        </div>
                    </div>
                </div>
            </div>

            {/* Configuration Types Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {configTypes.map((config) => {
                    const IconComponent = config.icon;
                    return (
                        <div
                            key={config.id}
                            onClick={() => handleConfigSelect(config)}
                            className={`group relative bg-white rounded-3xl border-2 transition-all duration-300 hover:shadow-lg cursor-pointer p-8 ${selectedConfig === config.id ? 'border-blue-500 shadow-lg' : 'border-gray-200 hover:border-gray-300'
                                }`}
                        >
                            <div className="flex items-start justify-between mb-6">
                                <div className={`w-14 h-14 ${config.color} rounded-2xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300`}>
                                    <IconComponent className="h-7 w-7 text-white" />
                                </div>
                                <div className="opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                                    <ArrowRight className="h-5 w-5 text-gray-400" />
                                </div>
                            </div>
                            <h3 className="text-xl font-bold text-gray-900 mb-3">
                                {config.name}
                            </h3>
                            <p className="text-gray-600 leading-relaxed">
                                {config.description}
                            </p>
                        </div>
                    );
                })}
            </div>
        </div>
    );
} 