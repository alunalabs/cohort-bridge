'use client';

import { useState } from 'react';
import { ArrowLeft, Download, Copy, Check, Settings as SettingsIcon } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useForm, FormProvider } from 'react-hook-form';
import yaml from 'js-yaml';

import DatabaseSection from './sections/DatabaseSection';
import PeerSection from './sections/PeerSection';
import SecuritySection from './sections/SecuritySection';
import TimeoutsSection from './sections/TimeoutsSection';
import LoggingSection from './sections/LoggingSection';
import MatchingSection from './sections/MatchingSection';

interface ConfigSection {
    id: string;
    name: string;
    component: React.ComponentType<any>;
    enabled: boolean;
}

interface FlexibleConfigBuilderProps {
    title: string;
    description: string;
    defaultSections: string[];
    defaultValues: any;
    icon?: React.ComponentType<any>;
}

export default function FlexibleConfigBuilder({
    title,
    description,
    defaultSections,
    defaultValues,
    icon: Icon = SettingsIcon
}: FlexibleConfigBuilderProps) {
    const router = useRouter();
    const [generatedYaml, setGeneratedYaml] = useState<string>('');
    const [copied, setCopied] = useState(false);

    const allSections: ConfigSection[] = [
        { id: 'database', name: 'Database Configuration', component: DatabaseSection, enabled: defaultSections.includes('database') },
        { id: 'peer', name: 'Peer Configuration', component: PeerSection, enabled: defaultSections.includes('peer') },
        { id: 'security', name: 'Security Configuration', component: SecuritySection, enabled: defaultSections.includes('security') },
        { id: 'timeouts', name: 'Timeouts Configuration', component: TimeoutsSection, enabled: defaultSections.includes('timeouts') },
        { id: 'logging', name: 'Logging Configuration', component: LoggingSection, enabled: defaultSections.includes('logging') },
        { id: 'matching', name: 'Matching Configuration', component: MatchingSection, enabled: defaultSections.includes('matching') },
    ];

    const [sections, setSections] = useState<ConfigSection[]>(allSections);
    const [validationError, setValidationError] = useState<string>('');

    const methods = useForm({
        defaultValues: defaultValues as any,
    });

    const toggleSection = (sectionId: string) => {
        const hasEnabledSections = sections.some(s => s.enabled);
        const targetSection = sections.find(s => s.id === sectionId);

        // Prevent disabling the last enabled section
        if (hasEnabledSections && sections.filter(s => s.enabled).length === 1 && targetSection?.enabled) {
            setValidationError('At least one configuration section must be enabled.');
            setTimeout(() => setValidationError(''), 3000);
            return;
        }

        setSections(prev => prev.map(section =>
            section.id === sectionId
                ? { ...section, enabled: !section.enabled }
                : section
        ));
        setValidationError('');
    };

    const validateRequiredFields = (data: any, enabledSectionIds: string[]) => {
        const errors: string[] = [];

        enabledSectionIds.forEach(sectionId => {
            if (sectionId === 'database' && data.database) {
                if (data.database.type === 'postgres') {
                    if (!data.database.host) errors.push('PostgreSQL host is required');
                    if (!data.database.port) errors.push('PostgreSQL port is required');
                    if (!data.database.user) errors.push('PostgreSQL username is required');
                    if (!data.database.dbname) errors.push('PostgreSQL database name is required');
                    if (!data.database.table) errors.push('PostgreSQL table name is required');
                } else if (data.database.type === 'csv') {
                    if (!data.database.filename) errors.push('CSV filename is required');
                }
            }

            if (sectionId === 'peer' && data.peer) {
                if (!data.peer.address) errors.push('Peer address is required');
                if (!data.peer.port) errors.push('Peer port is required');
            }
        });

        return errors;
    };

    const onSubmit = (data: any) => {
        try {
            // Clean up undefined values and filter based on enabled sections
            const enabledSectionIds = sections.filter(s => s.enabled).map(s => s.id);

            if (enabledSectionIds.length === 0) {
                setValidationError('At least one configuration section must be enabled.');
                return;
            }

            // Validate required fields
            const validationErrors = validateRequiredFields(data, enabledSectionIds);
            if (validationErrors.length > 0) {
                setValidationError(validationErrors.join('; '));
                return;
            }

            setValidationError('');
            const filteredData: any = {};

            // Always include basic fields
            if (data.listen_port !== undefined) filteredData.listen_port = data.listen_port;
            if (data.private_key !== undefined && data.private_key !== '') filteredData.private_key = data.private_key;
            if (data.public_key !== undefined && data.public_key !== '') filteredData.public_key = data.public_key;

            // Include section data based on what's enabled
            enabledSectionIds.forEach(sectionId => {
                if (data[sectionId] && Object.keys(data[sectionId]).length > 0) {
                    filteredData[sectionId] = data[sectionId];
                }
            });

            const yamlString = yaml.dump(filteredData, {
                indent: 2,
                lineWidth: 120,
                quotingType: '"',
                forceQuotes: false,
            });

            setGeneratedYaml(yamlString);
        } catch (error) {
            console.error('Error generating YAML:', error);
        }
    };

    const downloadYaml = () => {
        const blob = new Blob([generatedYaml], { type: 'text/yaml' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${title.toLowerCase().replace(/\s+/g, '_')}_config.yaml`;
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

    return (
        <>
            {/* Header */}
            <header className="bg-white shadow-sm border-b border-slate-200">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex justify-between items-center py-6">
                        <div className="flex items-center space-x-4">
                            <button
                                onClick={() => router.back()}
                                className="flex items-center space-x-2 text-slate-600 hover:text-slate-900 transition-colors cursor-pointer"
                            >
                                <ArrowLeft className="h-5 w-5" />
                                <span>Back</span>
                            </button>
                            <div className="w-px h-6 bg-slate-300" />
                            <div className="flex items-center space-x-3">
                                <Icon className="h-6 w-6 text-slate-700" />
                                <div>
                                    <h1 className="text-2xl font-bold text-slate-900">{title}</h1>
                                    <p className="text-sm text-slate-600">{description}</p>
                                </div>
                            </div>
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
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                    {/* Configuration Sections Sidebar */}
                    <div className="lg:col-span-1 space-y-4">
                        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
                            <h3 className="text-lg font-semibold text-slate-900 mb-4">Configuration Sections</h3>
                            <div className="space-y-3">
                                {sections.map((section) => (
                                    <div key={section.id}
                                        onClick={() => toggleSection(section.id)}
                                        className="flex items-center justify-between p-3 rounded-lg border border-slate-200 hover:border-slate-300 cursor-pointer transition-all duration-200">
                                        <span className="text-sm font-medium text-slate-700">{section.name}</span>
                                        <div className={`w-12 h-6 rounded-full transition-colors duration-200 ${section.enabled ? 'bg-blue-600' : 'bg-slate-300'
                                            }`}>
                                            <div className={`w-5 h-5 mt-0.5 bg-white rounded-full shadow-md transform transition-transform duration-200 ${section.enabled ? 'translate-x-6' : 'translate-x-0.5'
                                                }`}></div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>

                        {/* Preview Section */}
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
                                    <div className="text-center py-8 text-slate-500">
                                        <p className="text-sm font-medium mb-2">Preview will appear here</p>
                                        <p className="text-xs">Configure sections and generate to see YAML output</p>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>

                    {/* Form Section */}
                    <div className="lg:col-span-2 space-y-6">
                        <FormProvider {...methods}>
                            <form onSubmit={methods.handleSubmit(onSubmit)} className="space-y-6">
                                {sections.filter(section => section.enabled).map((section) => {
                                    const Component = section.component;
                                    return <Component key={section.id} configType={title.toLowerCase()} />;
                                })}

                                {validationError && (
                                    <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
                                        <div className="flex items-center">
                                            <div className="w-4 h-4 bg-red-500 rounded-full mr-3"></div>
                                            <p className="text-red-700 text-sm font-medium">{validationError}</p>
                                        </div>
                                    </div>
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
                </div>
            </div>
        </>
    );
} 