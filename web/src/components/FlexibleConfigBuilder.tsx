'use client';

import React, { useState, useCallback, useRef, useEffect } from 'react';
import { ArrowLeft, Copy, Check, Settings as SettingsIcon, Upload, Save } from 'lucide-react';
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
    const [missingFields, setMissingFields] = useState<string[]>([]);
    const [filename, setFilename] = useState(`config_${title.toLowerCase().replace(/\s+/g, '_')}.yaml`);
    const [showImportWarning, setShowImportWarning] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);

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

    // Watch for form changes to update preview in real-time
    const watchedValues = methods.watch();

    // Update preview when form values or sections change
    useEffect(() => {
        // Debounce the generation to avoid excessive updates
        const timeoutId = setTimeout(() => {
            generateConfiguration();
        }, 300);

        return () => clearTimeout(timeoutId);
    }, [watchedValues, sections]);

    // Check for imported configuration on component mount
    useEffect(() => {
        const isImporting = sessionStorage.getItem('isImporting');
        const importedConfigData = sessionStorage.getItem('importedConfig');

        if (isImporting === 'true' && importedConfigData) {
            try {
                const parsedConfig = JSON.parse(importedConfigData);

                // Enable sections based on what's in the imported config
                const sectionsToEnable: string[] = [];
                if (parsedConfig.database) sectionsToEnable.push('database');
                if (parsedConfig.peer) sectionsToEnable.push('peer');
                if (parsedConfig.security) sectionsToEnable.push('security');
                if (parsedConfig.timeouts) sectionsToEnable.push('timeouts');
                if (parsedConfig.logging) sectionsToEnable.push('logging');
                if (parsedConfig.matching) sectionsToEnable.push('matching');

                // Update sections state
                setSections(prev => prev.map(section => ({
                    ...section,
                    enabled: sectionsToEnable.includes(section.id)
                })));

                // Reset form with imported data
                methods.reset(parsedConfig);

                // Clear session storage
                sessionStorage.removeItem('isImporting');
                sessionStorage.removeItem('importedConfig');

                // Auto-generate preview with imported data after sections are updated
                setTimeout(() => {
                    generateConfiguration();
                }, 200);

            } catch (error) {
                console.error('Error processing imported config:', error);
                // Clear session storage on error
                sessionStorage.removeItem('isImporting');
                sessionStorage.removeItem('importedConfig');
            }
        }
    }, []); // Empty dependency array to run only on mount

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
        setMissingFields([]); // Clear missing fields when sections change
    };

    const validateRequiredFields = useCallback((data: any, enabledSectionIds: string[]) => {
        const errors: string[] = [];
        const missing: string[] = [];

        enabledSectionIds.forEach(sectionId => {
            if (sectionId === 'database' && data.database) {
                if (data.database.type === 'postgres') {
                    if (!data.database.host || data.database.host.trim() === '') {
                        errors.push('PostgreSQL host is required');
                        missing.push('database.host');
                    }
                    if (!data.database.port) {
                        errors.push('PostgreSQL port is required');
                        missing.push('database.port');
                    }
                    if (!data.database.user || data.database.user.trim() === '') {
                        errors.push('PostgreSQL username is required');
                        missing.push('database.user');
                    }
                    if (!data.database.dbname || data.database.dbname.trim() === '') {
                        errors.push('PostgreSQL database name is required');
                        missing.push('database.dbname');
                    }
                    if (!data.database.table || data.database.table.trim() === '') {
                        errors.push('PostgreSQL table name is required');
                        missing.push('database.table');
                    }
                } else if (data.database.type === 'csv' && !data.database.is_tokenized) {
                    if (!data.database.filename || data.database.filename.trim() === '') {
                        errors.push('CSV filename is required');
                        missing.push('database.filename');
                    }
                }
            }

            if (sectionId === 'peer' && data.peer) {
                if (!data.peer.host || data.peer.host.trim() === '') {
                    errors.push('Peer host is required');
                    missing.push('peer.host');
                }
                if (!data.peer.port) {
                    errors.push('Peer port is required');
                    missing.push('peer.port');
                }
            }

            // Add more validation for other sections as needed
            if (sectionId === 'security' && data.security) {
                if (data.security.tls_enabled && (!data.security.cert_file || data.security.cert_file.trim() === '')) {
                    errors.push('Certificate file is required when TLS is enabled');
                    missing.push('security.cert_file');
                }
                if (data.security.tls_enabled && (!data.security.key_file || data.security.key_file.trim() === '')) {
                    errors.push('Key file is required when TLS is enabled');
                    missing.push('security.key_file');
                }
            }
        });

        return { errors, missing };
    }, []);

    const generateConfiguration = () => {
        const originalData = methods.getValues();

        try {
            // Create a clean copy of the data to avoid modifying the original
            const data = JSON.parse(JSON.stringify(originalData));

            // Clean up undefined values and filter based on enabled sections
            const enabledSectionIds = sections.filter(s => s.enabled).map(s => s.id);

            if (enabledSectionIds.length === 0) {
                setValidationError('At least one configuration section must be enabled.');
                setMissingFields([]);
                setGeneratedYaml('');
                return;
            }

            // Validate required fields (use original data for validation)
            const validation = validateRequiredFields(originalData, enabledSectionIds);
            if (validation.errors.length > 0) {
                setValidationError(validation.errors.join('; '));
                setMissingFields(validation.missing);
                setGeneratedYaml('');
                return;
            }

            setValidationError('');
            setMissingFields([]);
            const filteredData: any = {};

            // Always include basic fields
            if (data.listen_port !== undefined) filteredData.listen_port = data.listen_port;



            // Include section data based on what's enabled
            enabledSectionIds.forEach(sectionId => {
                if (data[sectionId] && Object.keys(data[sectionId]).length > 0) {
                    const sectionData = { ...data[sectionId] };

                    // Special handling for database section to filter out irrelevant fields
                    if (sectionId === 'database') {
                        // Check if encryption is enabled from original data (UI state)
                        const isEncrypted = originalData._ui_is_encrypted;

                        // Remove encryption fields if encryption is not enabled OR if both fields are empty
                        if (!isEncrypted || (!sectionData.encryption_key && !sectionData.encryption_key_file)) {
                            delete sectionData.encryption_key;
                            delete sectionData.encryption_key_file;
                        }

                        // Filter fields based on database type
                        const dbType = sectionData.type;
                        if (dbType === 'csv') {
                            // For CSV, only keep relevant fields
                            const csvFields = ['type', 'filename', 'fields', 'random_bits_percent', 'is_tokenized'];
                            if (isEncrypted) {
                                csvFields.push('encryption_key', 'encryption_key_file');
                            }
                            Object.keys(sectionData).forEach(key => {
                                if (!csvFields.includes(key)) {
                                    delete sectionData[key];
                                }
                            });
                        } else if (dbType === 'postgres') {
                            // For PostgreSQL, remove CSV-specific fields and empty fields
                            delete sectionData.filename;
                            // Remove empty/null PostgreSQL fields
                            if (!sectionData.host || sectionData.host === '') delete sectionData.host;
                            if (!sectionData.port || sectionData.port === null) delete sectionData.port;
                            if (!sectionData.user || sectionData.user === '') delete sectionData.user;
                            if (!sectionData.password || sectionData.password === '') delete sectionData.password;
                            if (!sectionData.dbname || sectionData.dbname === '') delete sectionData.dbname;
                            if (!sectionData.table || sectionData.table === '') delete sectionData.table;
                        }

                        // Always remove empty/null fields
                        Object.keys(sectionData).forEach(key => {
                            if (sectionData[key] === null || sectionData[key] === '' || sectionData[key] === undefined) {
                                delete sectionData[key];
                            }
                        });
                    }

                    filteredData[sectionId] = sectionData;
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
            setGeneratedYaml('');
        }
    };

    const downloadYaml = () => {
        // Generate preview before downloading
        generateConfiguration();

        // Small delay to ensure generation completes
        setTimeout(() => {
            // Use File System Access API if available, otherwise fallback to download
            if ('showSaveFilePicker' in window) {
                try {
                    saveWithFilePicker();
                } catch (error) {
                    // Fallback to regular download
                    saveToDownloads();
                }
            } else {
                saveToDownloads();
            }
        }, 100);
    };

    const saveWithFilePicker = async () => {
        try {
            const fileHandle = await (window as any).showSaveFilePicker({
                suggestedName: filename,
                types: [{
                    description: 'YAML files',
                    accept: { 'text/yaml': ['.yaml', '.yml'] }
                }]
            });

            const writable = await fileHandle.createWritable();
            await writable.write(generatedYaml);
            await writable.close();
        } catch (error) {
            if ((error as any).name !== 'AbortError') {
                saveToDownloads();
            }
        }
    };

    const saveToDownloads = () => {
        const blob = new Blob([generatedYaml], { type: 'text/yaml' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    };

    const copyToClipboard = async () => {
        // Generate preview before copying
        generateConfiguration();

        // Small delay to ensure generation completes
        setTimeout(async () => {
            try {
                await navigator.clipboard.writeText(generatedYaml);
                setCopied(true);
                setTimeout(() => setCopied(false), 2000);
            } catch (error) {
                console.error('Failed to copy:', error);
            }
        }, 100);
    };

    const handleImportConfig = () => {
        // Check if there's existing content to warn about
        const currentData = methods.getValues();
        const hasContent = Object.keys(currentData).some(key => {
            const value = currentData[key];
            if (typeof value === 'object' && value !== null) {
                return Object.keys(value).length > 0;
            }
            return value !== undefined && value !== '';
        });

        if (hasContent) {
            setShowImportWarning(true);
        } else {
            fileInputRef.current?.click();
        }
    };

    const confirmImport = () => {
        setShowImportWarning(false);
        fileInputRef.current?.click();
    };

    const handleFileImport = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (!file) return;

        try {
            const text = await file.text();
            const parsedConfig = yaml.load(text) as any;

            if (!parsedConfig || typeof parsedConfig !== 'object') {
                setValidationError('Invalid YAML file format');
                return;
            }

            // Set filename from imported file
            setFilename(file.name);

            // Enable sections based on what's in the config
            const sectionsToEnable: string[] = [];
            if (parsedConfig.database) sectionsToEnable.push('database');
            if (parsedConfig.peer) sectionsToEnable.push('peer');
            if (parsedConfig.security) sectionsToEnable.push('security');
            if (parsedConfig.timeouts) sectionsToEnable.push('timeouts');
            if (parsedConfig.logging) sectionsToEnable.push('logging');
            if (parsedConfig.matching) sectionsToEnable.push('matching');

            // Update sections state
            setSections(prev => prev.map(section => ({
                ...section,
                enabled: sectionsToEnable.includes(section.id)
            })));

            // Reset form with imported data
            methods.reset(parsedConfig);

            // Clear any previous errors
            setValidationError('');
            setMissingFields([]);

            // Auto-generate preview with imported data
            setTimeout(() => {
                generateConfiguration();
            }, 200);

        } catch (error) {
            console.error('Error importing config:', error);
            setValidationError('Failed to parse YAML file. Please check the file format.');
        }

        // Reset file input
        if (fileInputRef.current) {
            fileInputRef.current.value = '';
        }
    };

    return (
        <>
            {/* Hidden file input */}
            <input
                ref={fileInputRef}
                type="file"
                accept=".yaml,.yml"
                onChange={handleFileImport}
                style={{ display: 'none' }}
            />

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
                        <div className="flex items-center space-x-3">
                            <button
                                onClick={handleImportConfig}
                                className="flex items-center space-x-2 px-4 py-2 rounded-lg text-slate-600 hover:text-slate-900 hover:bg-slate-100 transition-colors"
                            >
                                <Upload className="h-4 w-4" />
                                <span>Import</span>
                            </button>
                            <button
                                onClick={copyToClipboard}
                                disabled={!generatedYaml}
                                className={`flex items-center space-x-2 px-4 py-2 rounded-lg transition-colors ${generatedYaml
                                    ? 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
                                    : 'text-slate-400 cursor-not-allowed'
                                    }`}
                            >
                                {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                <span>{copied ? 'Copied!' : 'Copy'}</span>
                            </button>
                            <button
                                onClick={downloadYaml}
                                disabled={!generatedYaml}
                                className={`flex items-center space-x-2 px-4 py-2 rounded-lg transition-colors ${generatedYaml
                                    ? 'bg-blue-600 text-white hover:bg-blue-700'
                                    : 'bg-slate-300 text-slate-500 cursor-not-allowed'
                                    }`}
                            >
                                <Save className="h-4 w-4" />
                                <span>Save</span>
                            </button>
                        </div>
                    </div>
                </div>
            </header>

            <div className={`max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 ${showImportWarning ? 'pt-24' : ''}`}>
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
                                <div className="flex items-center justify-between">
                                    <div>
                                        <h3 className="text-lg font-semibold text-slate-900">Configuration Preview</h3>
                                        <p className="text-sm text-slate-600 mt-1">Updates automatically as you type</p>
                                    </div>
                                    <button
                                        onClick={copyToClipboard}
                                        disabled={!generatedYaml}
                                        className={`flex items-center space-x-2 px-3 py-1.5 rounded-lg text-sm transition-colors ${generatedYaml
                                            ? 'text-slate-600 hover:text-slate-900 hover:bg-slate-200 border border-slate-300'
                                            : 'text-slate-400 cursor-not-allowed border border-slate-200'
                                            }`}
                                        title="Copy configuration to clipboard"
                                    >
                                        {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                                        <span>{copied ? 'Copied!' : 'Copy'}</span>
                                    </button>
                                </div>
                            </div>
                            <div className="p-6">
                                {generatedYaml ? (
                                    <pre className="text-sm text-slate-700 whitespace-pre-wrap font-mono bg-slate-50 p-4 rounded-lg overflow-auto max-h-96">
                                        {generatedYaml}
                                    </pre>
                                ) : (
                                    <div className="text-center py-8 text-slate-500">
                                        <p className="text-sm font-medium mb-2">Configuration will appear here</p>
                                        <p className="text-xs">
                                            Configure your sections and click "Update Preview" to see the YAML output
                                        </p>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>

                    {/* Form Section */}
                    <div className="lg:col-span-2 space-y-6">
                        <FormProvider {...methods}>
                            <div className="space-y-6">
                                {/* Filename Section - Always visible */}
                                <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
                                    <div className="flex items-center space-x-3 mb-6">
                                        <Save className="h-5 w-5 text-slate-700" />
                                        <h3 className="text-lg font-semibold text-slate-900">File Configuration</h3>
                                    </div>
                                    <div className="space-y-4">
                                        <div>
                                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                                Filename
                                            </label>
                                            <input
                                                type="text"
                                                value={filename}
                                                onChange={(e) => setFilename(e.target.value)}
                                                className="w-full px-4 py-3 border border-slate-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                                placeholder="config.yaml"
                                            />
                                            <p className="mt-2 text-sm text-slate-500">
                                                This will be the default filename when saving your configuration.
                                            </p>
                                        </div>
                                    </div>
                                </div>

                                {sections.filter(section => section.enabled).map((section) => {
                                    const Component = section.component;
                                    return <Component
                                        key={section.id}
                                        configType={title.toLowerCase()}
                                        missingFields={missingFields}
                                    />;
                                })}

                                {validationError && (
                                    <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                                        <div className="flex items-center">
                                            <div className="w-4 h-4 bg-red-500 rounded-full mr-3"></div>
                                            <div>
                                                <p className="text-red-700 text-sm font-medium">{validationError}</p>
                                                {missingFields.length > 0 && (
                                                    <p className="text-red-600 text-xs mt-1">
                                                        Required fields are highlighted below.
                                                    </p>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </FormProvider>
                    </div>
                </div>
            </div>

            {/* Import Warning Banner */}
            {showImportWarning && (
                <div className="fixed top-0 left-0 right-0 z-50 bg-yellow-50 border-b border-yellow-200 p-4">
                    <div className="max-w-7xl mx-auto flex items-center justify-between">
                        <div className="flex items-center space-x-3">
                            <div className="w-8 h-8 bg-yellow-100 rounded-full flex items-center justify-center">
                                <Upload className="h-4 w-4 text-yellow-600" />
                            </div>
                            <div>
                                <h3 className="font-medium text-yellow-800">Import Configuration</h3>
                                <p className="text-sm text-yellow-700">
                                    Importing will replace all current settings. Any unsaved changes will be lost.
                                </p>
                            </div>
                        </div>
                        <div className="flex items-center space-x-3">
                            <button
                                onClick={() => setShowImportWarning(false)}
                                className="px-4 py-2 text-sm border border-yellow-300 text-yellow-700 rounded-lg hover:bg-yellow-100 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={confirmImport}
                                className="px-4 py-2 text-sm bg-yellow-500 text-white rounded-lg hover:bg-yellow-600 transition-colors"
                            >
                                Continue Import
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </>
    );
} 