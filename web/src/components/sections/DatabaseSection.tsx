'use client';

import { useFormContext } from 'react-hook-form';
import { Database, Plus, Minus, Settings, Info } from 'lucide-react';

interface DatabaseSectionProps {
    configType: string;
    missingFields?: string[];
}

export default function DatabaseSection({ configType, missingFields = [] }: DatabaseSectionProps) {
    const { register, watch, setValue } = useFormContext();

    const databaseType = watch('database.type');
    const fields = watch('database.fields') || [];
    const normalization = watch('database.normalization') || [];
    const isTokenized = watch('database.is_tokenized');

    const addField = () => {
        setValue('database.fields', [...fields, '']);
    };

    const removeField = (index: number) => {
        const newFields = fields.filter((_: any, i: number) => i !== index);
        setValue('database.fields', newFields);
    };

    const updateField = (index: number, value: string) => {
        const newFields = [...fields];
        newFields[index] = value;
        setValue('database.fields', newFields);
    };

    // Normalization functions
    const addNormalizationRule = () => {
        setValue('database.normalization', [...normalization, '']);
    };

    const removeNormalizationRule = (index: number) => {
        const newNormalization = normalization.filter((_: string, i: number) => i !== index);
        setValue('database.normalization', newNormalization);
    };

    const updateNormalizationRule = (index: number, value: string) => {
        const newNormalization = [...normalization];
        newNormalization[index] = value;
        setValue('database.normalization', newNormalization);
    };

    const normalizationMethods = [
        { value: 'name', label: 'Name', description: 'Lowercase, remove punctuation, normalize spaces' },
        { value: 'date', label: 'Date', description: 'Standardize to YYYY-MM-DD format' },
        { value: 'gender', label: 'Gender', description: 'Standardize to single characters (m/f/nb/o/u)' },
        { value: 'zip', label: 'ZIP', description: 'Extract first 5 digits, remove non-numeric' }
    ];

    const parseNormalizationRule = (rule: string) => {
        const parts = rule.split(':');
        return {
            method: parts[0] || '',
            field: parts[1] || ''
        };
    };

    const createNormalizationRule = (method: string, field: string) => {
        return method && field ? `${method}:${field}` : '';
    };

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
                    Configure your data source, field mappings, and normalization for privacy-preserving record linkage.
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
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                            Tokenized File Path
                        </label>
                        <input
                            type="text"
                            {...register('database.tokenized_file')}
                            placeholder="out/tokens_party_a.json"
                            className={getInputClass('database.tokenized_file')}
                        />
                    </div>
                )}

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

                {/* Fields Configuration */}
                {!isTokenized && (
                    <div>
                        <div className="flex items-center justify-between mb-3">
                            <div>
                                <label className="block text-sm font-medium text-slate-700">
                                    Database Field Mapping
                                </label>
                                <p className="text-sm text-slate-600 mt-1">
                                    Map your database column names to data types for record linkage
                                </p>
                            </div>
                            <button
                                type="button"
                                onClick={addField}
                                className="flex items-center space-x-1 text-blue-600 hover:text-blue-700 text-sm font-medium"
                            >
                                <Plus className="h-4 w-4" />
                                <span>Add Field</span>
                            </button>
                        </div>

                        <div className="space-y-3">
                            {fields.map((field: string, index: number) => {
                                const fieldLabels = [
                                    'First Name Column',
                                    'Last Name Column',
                                    'Date of Birth Column',
                                    'Gender Column',
                                    'ZIP Code Column',
                                    'Email Column',
                                    'ID Column',
                                    'Phone Column'
                                ];

                                return (
                                    <div key={index} className="flex items-center space-x-3">
                                        <div className="flex-1">
                                            <label className="block text-xs font-medium text-slate-600 mb-1">
                                                {fieldLabels[index] || `Field ${index + 1} Column`}
                                            </label>
                                            <input
                                                type="text"
                                                value={field}
                                                onChange={(e) => updateField(index, e.target.value)}
                                                placeholder={`${fieldLabels[index]?.split(' ')[0].toLowerCase() || 'field'}_name`}
                                                className={getInputClass(`database.fields.${index}`)}
                                            />
                                        </div>
                                        {fields.length > 1 && (
                                            <button
                                                type="button"
                                                onClick={() => removeField(index)}
                                                className="mt-6 p-2 text-red-600 hover:text-red-700 hover:bg-red-50 rounded-lg transition-colors"
                                            >
                                                <Minus className="h-4 w-4" />
                                            </button>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                )}

                {/* Data Normalization Section */}
                {!isTokenized && fields.length > 0 && (
                    <div className="border-t border-slate-200 pt-6">
                        <div className="mb-6">
                            <div className="flex items-center space-x-3 mb-2">
                                <div className="w-6 h-6 bg-purple-100 rounded-lg flex items-center justify-center">
                                    <Settings className="h-3 w-3 text-purple-600" />
                                </div>
                                <h4 className="text-md font-medium text-slate-900">Data Normalization</h4>
                            </div>
                            <p className="text-sm text-slate-600 ml-9">
                                Configure automatic data normalization to improve matching accuracy. Fields are standardized before tokenization.
                            </p>
                        </div>

                        {/* Info Box */}
                        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
                            <div className="flex items-start space-x-3">
                                <Info className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
                                <div>
                                    <h5 className="text-sm font-medium text-blue-900 mb-2">Normalization Benefits</h5>
                                    <ul className="text-sm text-blue-800 space-y-1">
                                        <li>• <strong>Name:</strong> &quot;Mary-Jane O&apos;Connor&quot; matches &quot;MARYJANE OCONNOR&quot;</li>
                                        <li>• <strong>Date:</strong> &quot;12/25/2023&quot; matches &quot;2023-12-25&quot;</li>
                                        <li>• <strong>ZIP:</strong> &quot;12345-6789&quot; matches &quot;12345 6789&quot;</li>
                                        <li>• <strong>Gender:</strong> &quot;Female&quot; matches &quot;f&quot;</li>
                                    </ul>
                                </div>
                            </div>
                        </div>

                        <div>
                            <div className="flex items-center justify-between mb-4">
                                <div>
                                    <label className="block text-sm font-medium text-slate-700">
                                        Normalization Rules
                                    </label>
                                    <p className="text-sm text-slate-600 mt-1">
                                        Map your database fields to normalization methods for improved matching
                                    </p>
                                </div>
                                <button
                                    type="button"
                                    onClick={addNormalizationRule}
                                    className="flex items-center space-x-1 text-purple-600 hover:text-purple-700 text-sm font-medium"
                                >
                                    <Plus className="h-4 w-4" />
                                    <span>Add Rule</span>
                                </button>
                            </div>

                            {normalization.length === 0 && (
                                <div className="text-center py-8 border-2 border-dashed border-slate-300 rounded-lg">
                                    <Settings className="h-12 w-12 text-slate-400 mx-auto mb-3" />
                                    <p className="text-slate-600 mb-2">No normalization rules configured</p>
                                    <p className="text-sm text-slate-500 mb-4">
                                        Set up normalization for: <strong>First Name → Last Name → Date of Birth → Gender → ZIP Code</strong>
                                    </p>
                                    <div className="flex justify-center space-x-3">
                                        <button
                                            type="button"
                                            onClick={() => {
                                                // Auto-setup common normalizations in the preferred order
                                                const commonSetup = [];

                                                // 1. First Name
                                                const firstNameField = fields.find((f: string) =>
                                                    f.toLowerCase().includes('first') ||
                                                    (f.toLowerCase().includes('name') && f.toLowerCase().includes('first')) ||
                                                    f.toLowerCase() === 'first' ||
                                                    f.toLowerCase() === 'fname'
                                                );
                                                if (firstNameField) {
                                                    commonSetup.push(createNormalizationRule('name', firstNameField));
                                                }

                                                // 2. Last Name  
                                                const lastNameField = fields.find((f: string) =>
                                                    f.toLowerCase().includes('last') ||
                                                    (f.toLowerCase().includes('name') && f.toLowerCase().includes('last')) ||
                                                    f.toLowerCase() === 'last' ||
                                                    f.toLowerCase() === 'lname' ||
                                                    f.toLowerCase() === 'surname'
                                                );
                                                if (lastNameField) {
                                                    commonSetup.push(createNormalizationRule('name', lastNameField));
                                                }

                                                // 3. Date of Birth
                                                const dobField = fields.find((f: string) =>
                                                    f.toLowerCase().includes('birth') ||
                                                    f.toLowerCase().includes('dob') ||
                                                    f.toLowerCase().includes('date') ||
                                                    f.toLowerCase() === 'birthdate'
                                                );
                                                if (dobField) {
                                                    commonSetup.push(createNormalizationRule('date', dobField));
                                                }

                                                // 4. Gender
                                                const genderField = fields.find((f: string) =>
                                                    f.toLowerCase().includes('gender') ||
                                                    f.toLowerCase().includes('sex') ||
                                                    f.toLowerCase() === 'gender' ||
                                                    f.toLowerCase() === 'sex'
                                                );
                                                if (genderField) {
                                                    commonSetup.push(createNormalizationRule('gender', genderField));
                                                }

                                                // 5. ZIP Code
                                                const zipField = fields.find((f: string) =>
                                                    f.toLowerCase().includes('zip') ||
                                                    f.toLowerCase().includes('postal') ||
                                                    f.toLowerCase() === 'zip' ||
                                                    f.toLowerCase() === 'zipcode' ||
                                                    f.toLowerCase() === 'zip_code'
                                                );
                                                if (zipField) {
                                                    commonSetup.push(createNormalizationRule('zip', zipField));
                                                }

                                                if (commonSetup.length > 0) {
                                                    setValue('database.normalization', commonSetup);
                                                }
                                            }}
                                            className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-purple-500 to-blue-500 text-white rounded-lg text-sm font-medium hover:from-purple-600 hover:to-blue-600 transition-all shadow-sm"
                                        >
                                            <Settings className="h-4 w-4 mr-2" />
                                            Auto-Setup Common Fields
                                        </button>
                                        <button
                                            type="button"
                                            onClick={addNormalizationRule}
                                            className="inline-flex items-center px-4 py-2 border border-slate-300 text-slate-700 rounded-lg text-sm font-medium hover:border-slate-400 hover:text-slate-900 transition-colors bg-white"
                                        >
                                            <Plus className="h-4 w-4 mr-2" />
                                            Add Custom Rule
                                        </button>
                                    </div>
                                </div>
                            )}

                            {normalization.length > 0 && (
                                <div className="space-y-3">
                                    {normalization.map((rule: string, index: number) => {
                                        const parsed = parseNormalizationRule(rule);
                                        return (
                                            <div key={index} className="flex items-center space-x-3 p-4 bg-slate-50 rounded-lg">
                                                <div className="flex-1 grid grid-cols-2 gap-3">
                                                    <div>
                                                        <label className="block text-xs font-medium text-slate-600 mb-1">
                                                            Normalization Method
                                                        </label>
                                                        <select
                                                            value={parsed.method}
                                                            onChange={(e) => updateNormalizationRule(index, createNormalizationRule(e.target.value, parsed.field))}
                                                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 text-slate-900 bg-white"
                                                        >
                                                            <option value="">Select method...</option>
                                                            {normalizationMethods.map(method => (
                                                                <option key={method.value} value={method.value}>
                                                                    {method.label}
                                                                </option>
                                                            ))}
                                                        </select>
                                                    </div>
                                                    <div>
                                                        <label className="block text-xs font-medium text-slate-600 mb-1">
                                                            Database Field
                                                        </label>
                                                        <select
                                                            value={parsed.field}
                                                            onChange={(e) => updateNormalizationRule(index, createNormalizationRule(parsed.method, e.target.value))}
                                                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 text-slate-900 bg-white"
                                                        >
                                                            <option value="">Select field...</option>
                                                            {fields.map((field: string, fieldIndex: number) => (
                                                                <option key={fieldIndex} value={field}>
                                                                    {field}
                                                                </option>
                                                            ))}
                                                        </select>
                                                    </div>
                                                </div>
                                                <button
                                                    type="button"
                                                    onClick={() => removeNormalizationRule(index)}
                                                    className="p-2 text-red-600 hover:text-red-700 hover:bg-red-50 rounded-lg transition-colors"
                                                >
                                                    <Minus className="h-4 w-4" />
                                                </button>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
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