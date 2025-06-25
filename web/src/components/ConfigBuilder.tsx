'use client';

import { useState } from 'react';
import { ArrowLeft, Download, Copy, Check } from 'lucide-react';
import { useForm, FormProvider } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import yaml from 'js-yaml';

import DatabaseSection from './sections/DatabaseSection';
import PeerSection from './sections/PeerSection';
import SecuritySection from './sections/SecuritySection';
import TimeoutsSection from './sections/TimeoutsSection';
import LoggingSection from './sections/LoggingSection';
import MatchingSection from './sections/MatchingSection';

// Simplified schema - we'll handle validation in the form
const configSchema = z.object({
    database: z.any(),
    peer: z.any(),
    listen_port: z.any(),
    private_key: z.string().optional(),
    public_key: z.string().optional(),
    security: z.any().optional(),
    timeouts: z.any().optional(),
    logging: z.any().optional(),
    matching: z.any().optional(),
});

interface ConfigBuilderProps {
    configType: string;
    onBack: () => void;
}

export default function ConfigBuilder({ configType, onBack }: ConfigBuilderProps) {
    const [generatedYaml, setGeneratedYaml] = useState<string>('');
    const [copied, setCopied] = useState(false);
    const [showPreview, setShowPreview] = useState(false);

    // Use simplified schema for all types
    const getSchema = () => configSchema;

    // Get default values based on config type
    const getDefaultValues = () => {
        const base = {
            database: {
                type: 'csv' as const,
                fields: ['first_name', 'last_name', 'date_of_birth'],
                random_bits_percent: 0.0,
            },
            peer: {
                host: 'localhost',
                port: 8081,
            },
            listen_port: 8080,
        };

        switch (configType) {
            case 'postgres':
                return {
                    ...base,
                    database: {
                        ...base.database,
                        type: 'postgres' as const,
                        host: 'localhost',
                        port: 5432,
                        user: 'cohort_user',
                        password: '',
                        dbname: 'cohort_database',
                        table: 'users',
                    },
                };
            case 'secure':
                return {
                    ...base,
                    security: {
                        allowed_ips: ['127.0.0.1', '::1'],
                        require_ip_check: true,
                        max_connections: 5,
                        rate_limit_per_min: 10,
                    },
                    timeouts: {
                        connection_timeout: '30s',
                        read_timeout: '60s',
                        write_timeout: '60s',
                        idle_timeout: '300s',
                        handshake_timeout: '30s',
                    },
                    logging: {
                        level: 'info' as const,
                        file: 'logs/cohort.log',
                        max_size: 100,
                        max_backups: 3,
                        max_age: 30,
                        enable_audit: true,
                        audit_file: 'logs/audit.log',
                    },
                };
            case 'tokenized':
                return {
                    ...base,
                    database: {
                        ...base.database,
                        is_tokenized: true,
                        tokenized_file: 'out/tokens_party_a.json',
                    },
                };
            case 'basic':
            case 'network':
                return {
                    ...base,
                    matching: {
                        bloom_size: 2048,
                        bloom_hashes: 8,
                        minhash_size: 256,
                        qgram_length: 3,
                        hamming_threshold: 200,
                        jaccard_threshold: 0.75,
                        qgram_threshold: 0.85,
                        noise_level: 0.02,
                    },
                };
            default:
                return base;
        }
    };

    const methods = useForm({
        defaultValues: getDefaultValues() as any,
    });

    const onSubmit = (data: any) => {
        try {
            // Clean up undefined values
            const cleanData = JSON.parse(JSON.stringify(data, (key, value) => {
                return value === undefined ? null : value;
            }));

            const yamlString = yaml.dump(cleanData, {
                indent: 2,
                lineWidth: 120,
                quotingType: '"',
                forceQuotes: false,
            });

            setGeneratedYaml(yamlString);
            setShowPreview(true);
        } catch (error) {
            console.error('Error generating YAML:', error);
        }
    };

    const downloadYaml = () => {
        const blob = new Blob([generatedYaml], { type: 'text/yaml' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `config_${configType}.yaml`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    };

    const copyToClipboard = async () => {
        try {
            await navigator.clipboard.writeText(generatedYaml);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (error) {
            console.error('Failed to copy:', error);
        }
    };

    const getConfigTitle = () => {
        switch (configType) {
            case 'basic': return 'Basic Configuration';
            case 'postgres': return 'PostgreSQL Configuration';
            case 'secure': return 'Secure Configuration';
            case 'tokenized': return 'Tokenized Configuration';
            case 'network': return 'Network Configuration';
            default: return 'Configuration';
        }
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
            {/* Header */}
            <header className="bg-white shadow-sm border-b border-slate-200">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex justify-between items-center py-6">
                        <div className="flex items-center space-x-4">
                            <button
                                onClick={onBack}
                                className="flex items-center space-x-2 text-slate-600 hover:text-slate-900 transition-colors"
                            >
                                <ArrowLeft className="h-5 w-5" />
                                <span>Back</span>
                            </button>
                            <div className="w-px h-6 bg-slate-300" />
                            <h1 className="text-2xl font-bold text-slate-900">{getConfigTitle()}</h1>
                        </div>
                        {generatedYaml && (
                            <div className="flex items-center space-x-3">
                                <button
                                    onClick={copyToClipboard}
                                    className="flex items-center space-x-2 px-4 py-2 text-slate-600 hover:text-slate-900 hover:bg-slate-100 rounded-lg transition-colors"
                                >
                                    {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                    <span>{copied ? 'Copied!' : 'Copy'}</span>
                                </button>
                                <button
                                    onClick={downloadYaml}
                                    className="flex items-center space-x-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
                                >
                                    <Download className="h-4 w-4" />
                                    <span>Download</span>
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            </header>

            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                    {/* Form Section */}
                    <div className="space-y-6">
                        <FormProvider {...methods}>
                            <form onSubmit={methods.handleSubmit(onSubmit)} className="space-y-6">
                                <DatabaseSection configType={configType} />
                                <PeerSection />

                                {(configType === 'secure' || configType === 'tokenized') && (
                                    <SecuritySection />
                                )}

                                {(configType === 'secure' || configType === 'tokenized') && (
                                    <TimeoutsSection />
                                )}

                                {(configType === 'secure' || configType === 'tokenized') && (
                                    <LoggingSection />
                                )}

                                {(configType === 'basic' || configType === 'network') && (
                                    <MatchingSection />
                                )}

                                <div className="flex justify-center pt-6">
                                    <button
                                        type="submit"
                                        className="bg-blue-600 text-white px-8 py-3 rounded-lg hover:bg-blue-700 transition-colors font-medium"
                                    >
                                        Generate Configuration
                                    </button>
                                </div>
                            </form>
                        </FormProvider>
                    </div>

                    {/* Preview Section */}
                    <div className="lg:sticky lg:top-8 lg:h-fit">
                        <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
                            <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
                                <h3 className="text-lg font-semibold text-slate-900">Configuration Preview</h3>
                            </div>
                            <div className="p-6">
                                {generatedYaml ? (
                                    <pre className="text-sm text-slate-700 whitespace-pre-wrap font-mono bg-slate-50 p-4 rounded-lg overflow-auto max-h-96">
                                        {generatedYaml}
                                    </pre>
                                ) : (
                                    <div className="text-center py-12 text-slate-500">
                                        <p className="text-lg font-medium mb-2">Preview will appear here</p>
                                        <p className="text-sm">Fill out the form and click "Generate Configuration" to see your YAML output</p>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
} 