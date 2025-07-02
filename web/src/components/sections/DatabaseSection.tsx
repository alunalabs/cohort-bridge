'use client';

import { useFormContext } from 'react-hook-form';
import { Database, Plus, Minus, Info } from 'lucide-react';
import { useEffect } from 'react';

interface DatabaseSectionProps {
    configType: string;
    missingFields?: string[];
}

interface FieldMapping {
    column: string;
    normalization: string;
}

export default function DatabaseSection({ configType, missingFields = [] }: DatabaseSectionProps) {
    const { register, watch, setValue } = useFormContext();

    const databaseType = watch('database.type');
    const fieldMappings = watch('_ui_field_mappings') || [];
    const isTokenized = watch('database.is_tokenized');
    const isEncrypted = watch('_ui_is_encrypted') || false;

    // Initialize with default fields and normalizations on first load
    useEffect(() => {
        if (fieldMappings.length === 0 && !isTokenized) {
            const defaultMappings = [
                { column: 'first_name', normalization: 'name' },
                { column: 'last_name', normalization: 'name' },
                { column: 'date_of_birth', normalization: 'date' },
                { column: 'gender', normalization: 'gender' },
                { column: 'zip_code', normalization: 'zip' }
            ];
            setValue('_ui_field_mappings', defaultMappings);
            // Also update the new fields format
            updateFieldsFromMappings(defaultMappings);
        }
    }, [fieldMappings.length, isTokenized, setValue]);

    // Clear encryption fields when encryption is disabled
    useEffect(() => {
        if (!isEncrypted) {
            setValue('database.encryption_key', '');
            setValue('database.encryption_key_file', '');
        }
    }, [isEncrypted, setValue]);

    // Convert field mappings to the new fields format
    const updateFieldsFromMappings = (mappings: FieldMapping[]) => {
        const fields = mappings.map(mapping => {
            if (mapping.normalization && mapping.normalization !== '') {
                return `${mapping.normalization}:${mapping.column}`;
            }
            return mapping.column;
        }).filter(field => field); // Remove empty fields
        setValue('database.fields', fields);
    };

    const addFieldMapping = () => {
        const newMappings = [...fieldMappings, { column: '', normalization: '' }];
        setValue('_ui_field_mappings', newMappings);
        updateFieldsFromMappings(newMappings);
    };

    const removeFieldMapping = (index: number) => {
        const newMappings = fieldMappings.filter((_: any, i: number) => i !== index);
        setValue('_ui_field_mappings', newMappings);
        updateFieldsFromMappings(newMappings);
    };

    const updateFieldMapping = (index: number, field: 'column' | 'normalization', value: string) => {
        const newMappings = [...fieldMappings];
        newMappings[index] = { ...newMappings[index], [field]: value };
        setValue('_ui_field_mappings', newMappings);
        updateFieldsFromMappings(newMappings);
    };

    const normalizationMethods = [
        { value: '', label: 'No normalization', description: 'Use data as-is' },
        { value: 'name', label: 'Name', description: 'Lowercase, remove punctuation, normalize spaces' },
        { value: 'date', label: 'Date', description: 'Standardize to YYYY-MM-DD format' },
        { value: 'gender', label: 'Gender', description: 'Standardize to single characters (m/f/nb/o/u)' },
        { value: 'zip', label: 'ZIP Code', description: 'Extract first 5 digits, remove non-numeric' }
    ];

    const getInputClass = (fieldName: string) => {
        const baseClass = "w-full px-3 py-2 border rounded-lg focus:ring-2 text-slate-900 placeholder-slate-400 bg-white transition-colors";
        const isMissing = missingFields.includes(fieldName);

        if (isMissing) {
            return `${baseClass} border-red-300 bg-red-50 focus:ring-red-500 focus:border-red-500`;
        }

        return `${baseClass} border-slate-300 focus:ring-blue-500 focus:border-blue-500`;
    };

    return (
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
            <div className="mb-6">
                <div className="flex items-center space-x-3 mb-2">
                    <div className="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
                        <Database className="h-4 w-4 text-blue-600" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900">Database Configuration</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Configure your data source and field mappings with normalization for privacy-preserving record linkage.
                </p>
            </div>

            <div className="space-y-6">
                {/* Database Type */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-3">
                        Database Type
                    </label>
                    <div className="grid grid-cols-2 gap-3">
                        <button
                            type="button"
                            onClick={() => setValue('database.type', 'csv')}
                            className={`p-4 rounded-lg border-2 transition-all duration-200 cursor-pointer ${databaseType === 'csv'
                                ? 'border-blue-500 bg-blue-50 text-blue-700'
                                : 'border-slate-300 hover:border-slate-400 text-slate-700'
                                }`}
                        >
                            <div className="text-center">
                                <div className="text-lg font-semibold">CSV File</div>
                                <div className="text-xs text-slate-500 mt-1">Comma-separated values</div>
                            </div>
                        </button>
                        <button
                            type="button"
                            onClick={() => setValue('database.type', 'postgres')}
                            className={`p-4 rounded-lg border-2 transition-all duration-200 cursor-pointer ${databaseType === 'postgres'
                                ? 'border-blue-500 bg-blue-50 text-blue-700'
                                : 'border-slate-300 hover:border-slate-400 text-slate-700'
                                }`}
                        >
                            <div className="text-center">
                                <div className="text-lg font-semibold">PostgreSQL</div>
                                <div className="text-xs text-slate-500 mt-1">Database server</div>
                            </div>
                        </button>
                    </div>
                </div>

                {/* Tokenized Toggle for tokenized config type */}
                {configType === 'tokenized' && (
                    <div>
                        <label className="flex items-center space-x-3">
                            <input
                                type="checkbox"
                                {...register('database.is_tokenized')}
                                className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                            />
                            <span className="text-sm font-medium text-slate-700">Use tokenized data</span>
                        </label>
                    </div>
                )}

                {/* CSV File Configuration */}
                {databaseType === 'csv' && !isTokenized && (
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                            CSV Filename <span className="text-red-500">*</span>
                            {missingFields.includes('database.filename') && (
                                <span className="text-red-500 text-xs ml-2">(Required field missing)</span>
                            )}
                        </label>
                        <input
                            type="text"
                            {...register('database.filename')}
                            placeholder="data/patients.csv"
                            className={getInputClass('database.filename')}
                        />
                    </div>
                )}

                {/* Tokenized File Configuration */}
                {isTokenized && (
                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Tokenized File Path
                            </label>
                            <input
                                type="text"
                                {...register('database.filename')}
                                placeholder="out/tokens_party_a.json"
                                className={getInputClass('database.filename')}
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                Path to the tokenized data file (CSV or JSON format)
                            </p>
                        </div>
                    </div>
                )}

                {/* Encryption Configuration - Show for all database types */}
                <div className="border-t pt-4">
                    <h4 className="text-sm font-medium text-slate-700 mb-3">Encryption Configuration</h4>
                    <p className="text-xs text-slate-600 mb-4">
                        Configure encryption settings for tokenized data processing.
                    </p>

                    {/* Encryption Toggle */}
                    <div className="mb-4">
                        <label className="flex items-center space-x-3">
                            <input
                                type="checkbox"
                                {...register('_ui_is_encrypted')}
                                className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                            />
                            <span className="text-sm font-medium text-slate-700">My data is encrypted</span>
                        </label>
                        <p className="text-xs text-slate-600 ml-6 mt-1">
                            Check this if your data file was created with encryption enabled
                        </p>
                    </div>

                    {/* Encryption Key Fields - Only shown when encrypted is checked */}
                    {isEncrypted && (
                        <div className="space-y-3 pl-4 border-l-2 border-blue-200">
                            <p className="text-xs text-slate-600 mb-3">
                                Provide either a hex key directly or a path to a key file to decrypt your data.
                            </p>

                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Encryption Key (Hex)
                                </label>
                                <input
                                    type="text"
                                    {...register('database.encryption_key')}
                                    placeholder="64-character hex key (leave empty if using key file)"
                                    className={getInputClass('database.encryption_key') + ' font-mono text-sm'}
                                />
                                <p className="mt-1 text-xs text-slate-600">
                                    32-byte encryption key as a 64-character hex string
                                </p>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Encryption Key File
                                </label>
                                <input
                                    type="text"
                                    {...register('database.encryption_key_file')}
                                    placeholder="out/tokenized_data.key"
                                    className={getInputClass('database.encryption_key_file')}
                                />
                                <p className="mt-1 text-xs text-slate-600">
                                    Path to file containing the encryption key (alternative to direct key input)
                                </p>
                            </div>

                            <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 mt-3">
                                <div className="flex items-start space-x-2">
                                    <div className="text-blue-600 font-medium text-xs">🔑</div>
                                    <div className="text-xs text-blue-700">
                                        <strong>Note:</strong> You only need to provide either the hex key OR the key file path, not both.
                                        The system will automatically decrypt your data during processing.
                                    </div>
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Information when not encrypted */}
                    {!isEncrypted && (
                        <div className="bg-gray-50 border border-gray-200 rounded-lg p-3">
                            <div className="flex items-start space-x-2">
                                <div className="text-gray-600 font-medium text-xs">ℹ️</div>
                                <div className="text-xs text-gray-700">
                                    Your data will be processed as unencrypted. If your data was created with encryption,
                                    please check the box above and provide the decryption key.
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                {/* PostgreSQL Configuration */}
                {databaseType === 'postgres' && (
                    <div className="space-y-4">
                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Host <span className="text-red-500">*</span>
                                    {missingFields.includes('database.host') && (
                                        <span className="text-red-500 text-xs ml-2">(Required field missing)</span>
                                    )}
                                </label>
                                <input
                                    type="text"
                                    {...register('database.host')}
                                    placeholder="localhost"
                                    className={getInputClass('database.host')}
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Port <span className="text-red-500">*</span>
                                    {missingFields.includes('database.port') && (
                                        <span className="text-red-500 text-xs ml-2">(Required field missing)</span>
                                    )}
                                </label>
                                <input
                                    type="number"
                                    {...register('database.port', { valueAsNumber: true })}
                                    placeholder="5432"
                                    className={getInputClass('database.port')}
                                />
                            </div>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Username <span className="text-red-500">*</span>
                                    {missingFields.includes('database.user') && (
                                        <span className="text-red-500 text-xs ml-2">(Required field missing)</span>
                                    )}
                                </label>
                                <input
                                    type="text"
                                    {...register('database.user')}
                                    placeholder="cohort_user"
                                    className={getInputClass('database.user')}
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Password
                                </label>
                                <input
                                    type="password"
                                    {...register('database.password')}
                                    placeholder="your_password_here"
                                    className={getInputClass('database.password')}
                                />
                            </div>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Database Name <span className="text-red-500">*</span>
                                    {missingFields.includes('database.dbname') && (
                                        <span className="text-red-500 text-xs ml-2">(Required field missing)</span>
                                    )}
                                </label>
                                <input
                                    type="text"
                                    {...register('database.dbname')}
                                    placeholder="cohort_database"
                                    className={getInputClass('database.dbname')}
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-2">
                                    Table Name <span className="text-red-500">*</span>
                                    {missingFields.includes('database.table') && (
                                        <span className="text-red-500 text-xs ml-2">(Required field missing)</span>
                                    )}
                                </label>
                                <input
                                    type="text"
                                    {...register('database.table')}
                                    placeholder="users"
                                    className={getInputClass('database.table')}
                                />
                            </div>
                        </div>
                    </div>
                )}

                {/* Field Mapping with Normalization */}
                {!isTokenized && (
                    <div>
                        <div className="flex items-center justify-between mb-4">
                            <div>
                                <label className="block text-sm font-medium text-slate-700">
                                    Field Mapping & Normalization
                                </label>
                                <p className="text-sm text-slate-600 mt-1">
                                    Map database columns to normalization methods for improved matching
                                </p>
                            </div>
                            <button
                                type="button"
                                onClick={addFieldMapping}
                                className="flex items-center space-x-1 text-blue-600 hover:text-blue-700 text-sm font-medium"
                            >
                                <Plus className="h-4 w-4" />
                                <span>Add Field</span>
                            </button>
                        </div>

                        {/* Recommendation Note */}
                        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
                            <div className="flex items-start space-x-3">
                                <Info className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
                                <div>
                                    <h5 className="text-sm font-medium text-blue-900 mb-2">Recommended Token Order</h5>
                                    <p className="text-sm text-blue-800 mb-2">
                                        We recommend generating tokens in this order: <strong>first name, last name, date of birth, gender, ZIP code</strong>.
                                        You can customize the order however you like, though we recommend placing additional fields at the end.
                                    </p>
                                    <p className="text-xs text-blue-700">
                                        This order optimizes for the most common matching scenarios in healthcare record linkage.
                                    </p>
                                </div>
                            </div>
                        </div>

                        <div className="flex-1 grid grid-cols-2 pr-8">
                            <label className="block text-xs font-medium text-slate-600 mb-1">
                                Database Column Name
                            </label>
                            <label className="block text-xs font-medium text-slate-600 mb-1">
                                Normalization Method
                            </label>
                        </div>

                        <div className="space-y-3">
                            {fieldMappings.map((mapping: FieldMapping, index: number) => (
                                <div key={index} className="flex items-center space-x-3 bg-slate-50 rounded-lg">
                                    <div className="flex-1 grid grid-cols-2 gap-3">
                                        <input
                                            type="text"
                                            value={mapping.column}
                                            onChange={(e) => updateFieldMapping(index, 'column', e.target.value)}
                                            placeholder="column_name"
                                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 bg-white"
                                        />
                                        <select
                                            value={mapping.normalization}
                                            onChange={(e) => updateFieldMapping(index, 'normalization', e.target.value)}
                                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 bg-white"
                                        >
                                            {normalizationMethods.map(method => (
                                                <option key={method.value} value={method.value}>
                                                    {method.label}
                                                </option>
                                            ))}
                                        </select>
                                    </div>
                                    {fieldMappings.length > 1 && (
                                        <button
                                            type="button"
                                            onClick={() => removeFieldMapping(index)}
                                            className="p-2 text-red-600 hover:text-red-700 hover:bg-red-50 rounded-lg transition-colors"
                                        >
                                            <Minus className="h-4 w-4" />
                                        </button>
                                    )}
                                </div>
                            ))}
                        </div>

                        {/* Normalization Info */}
                        <div className="mt-4 bg-slate-50 border border-slate-200 rounded-lg p-4">
                            <h6 className="text-sm font-medium text-slate-900 mb-2">Normalization Methods</h6>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs text-slate-600">
                                <div>
                                    <strong>Name:</strong> Converts "Mary-Jane O'Connor" → "maryjane oconnor"
                                </div>
                                <div>
                                    <strong>Date:</strong> Converts "12/25/2023" → "2023-12-25"
                                </div>
                                <div>
                                    <strong>Gender:</strong> Converts "Female" → "f"
                                </div>
                                <div>
                                    <strong>ZIP Code:</strong> Converts "12345-6789" → "12345"
                                </div>
                            </div>
                        </div>
                    </div>
                )}

                {/* Random Bits Percent */}
                <div className="border-t border-slate-200 pt-6">
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Random Bits Percentage (0.0 - 1.0)
                    </label>
                    <input
                        type="number"
                        step="0.01"
                        min="0"
                        max="1"
                        {...register('database.random_bits_percent', { valueAsNumber: true })}
                        placeholder="0.0"
                        className={getInputClass('database.random_bits_percent')}
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Higher values (0.05-0.1) increase privacy but reduce matching accuracy. Use 0.0 for maximum accuracy.
                    </p>
                </div>
            </div>
        </div>
    );
} 