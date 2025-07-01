'use client';

import { useFormContext } from 'react-hook-form';
import { Settings, Plus, Minus, Info } from 'lucide-react';

interface NormalizationSectionProps {
    missingFields?: string[];
}

export default function NormalizationSection({ missingFields = [] }: NormalizationSectionProps) {
    const { watch, setValue } = useFormContext();

    const normalization = watch('database.normalization') || [];
    const databaseFields = watch('database.fields') || [];

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

    const getInputClass = (fieldName: string) => {
        const baseClass = "w-full px-3 py-2 border rounded-lg focus:ring-2 text-slate-900 placeholder-slate-400 bg-white transition-colors";
        const isMissing = missingFields.includes(fieldName);

        if (isMissing) {
            return `${baseClass} border-red-300 bg-red-50 focus:ring-red-500 focus:border-red-500`;
        }

        return `${baseClass} border-slate-300 focus:ring-blue-500 focus:border-blue-500`;
    };

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

    return (
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
            <div className="mb-6">
                <div className="flex items-center space-x-3 mb-2">
                    <div className="w-8 h-8 bg-purple-100 rounded-lg flex items-center justify-center">
                        <Settings className="h-4 w-4 text-purple-600" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900">Data Normalization</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Configure automatic data normalization to improve matching accuracy. Fields are standardized before tokenization.
                </p>
            </div>

            {/* Info Box */}
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
                <div className="flex items-start space-x-3">
                    <Info className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
                    <div>
                        <h4 className="text-sm font-medium text-blue-900 mb-2">Normalization Benefits</h4>
                        <ul className="text-sm text-blue-800 space-y-1">
                            <li>• <strong>Name:</strong> &quot;Mary-Jane O&apos;Connor&quot; matches &quot;MARYJANE OCONNOR&quot;</li>
                            <li>• <strong>Date:</strong> &quot;12/25/2023&quot; matches &quot;2023-12-25&quot;</li>
                            <li>• <strong>ZIP:</strong> &quot;12345-6789&quot; matches &quot;12345 6789&quot;</li>
                            <li>• <strong>Gender:</strong> &quot;Female&quot; matches &quot;f&quot;</li>
                        </ul>
                    </div>
                </div>
            </div>

            <div className="space-y-6">
                {/* Normalization Rules */}
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
                                        const firstNameField = databaseFields.find((f: string) =>
                                            f.toLowerCase().includes('first') ||
                                            (f.toLowerCase().includes('name') && f.toLowerCase().includes('first')) ||
                                            f.toLowerCase() === 'first' ||
                                            f.toLowerCase() === 'fname'
                                        );
                                        if (firstNameField) {
                                            commonSetup.push(createNormalizationRule('name', firstNameField));
                                        }

                                        // 2. Last Name  
                                        const lastNameField = databaseFields.find((f: string) =>
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
                                        const dobField = databaseFields.find((f: string) =>
                                            f.toLowerCase().includes('birth') ||
                                            f.toLowerCase().includes('dob') ||
                                            f.toLowerCase().includes('date') ||
                                            f.toLowerCase() === 'birthdate'
                                        );
                                        if (dobField) {
                                            commonSetup.push(createNormalizationRule('date', dobField));
                                        }

                                        // 4. Gender
                                        const genderField = databaseFields.find((f: string) =>
                                            f.toLowerCase().includes('gender') ||
                                            f.toLowerCase().includes('sex') ||
                                            f.toLowerCase() === 'gender' ||
                                            f.toLowerCase() === 'sex'
                                        );
                                        if (genderField) {
                                            commonSetup.push(createNormalizationRule('gender', genderField));
                                        }

                                        // 5. ZIP Code
                                        const zipField = databaseFields.find((f: string) =>
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

                    <div className="space-y-4">
                        {normalization.map((rule: string, index: number) => {
                            const { method, field } = parseNormalizationRule(rule);

                            return (
                                <div key={index} className="bg-slate-50 rounded-lg p-4 border border-slate-200">
                                    <div className="flex items-center space-x-4">
                                        <div className="flex-1 grid grid-cols-2 gap-4">
                                            {/* Normalization Method */}
                                            <div>
                                                <label className="block text-xs font-medium text-slate-600 mb-2">
                                                    Normalization Method
                                                </label>
                                                <select
                                                    value={method}
                                                    onChange={(e) => {
                                                        const newRule = createNormalizationRule(e.target.value, field);
                                                        updateNormalizationRule(index, newRule);
                                                    }}
                                                    className={getInputClass(`normalization.${index}.method`)}
                                                >
                                                    <option value="">Select method...</option>
                                                    {normalizationMethods.map((nm) => (
                                                        <option key={nm.value} value={nm.value}>
                                                            {nm.label}
                                                        </option>
                                                    ))}
                                                </select>
                                            </div>

                                            {/* Field Name */}
                                            <div>
                                                <label className="block text-xs font-medium text-slate-600 mb-2">
                                                    Database Field Name
                                                </label>
                                                {databaseFields.length > 0 ? (
                                                    <select
                                                        value={field}
                                                        onChange={(e) => {
                                                            const newRule = createNormalizationRule(method, e.target.value);
                                                            updateNormalizationRule(index, newRule);
                                                        }}
                                                        className={getInputClass(`normalization.${index}.field`)}
                                                    >
                                                        <option value="">Select field...</option>
                                                        {databaseFields.map((dbField: string, fieldIndex: number) => (
                                                            <option key={fieldIndex} value={dbField}>
                                                                {dbField}
                                                            </option>
                                                        ))}
                                                    </select>
                                                ) : (
                                                    <input
                                                        type="text"
                                                        value={field}
                                                        onChange={(e) => {
                                                            const newRule = createNormalizationRule(method, e.target.value);
                                                            updateNormalizationRule(index, newRule);
                                                        }}
                                                        placeholder="Enter field name..."
                                                        className={getInputClass(`normalization.${index}.field`)}
                                                    />
                                                )}
                                            </div>
                                        </div>

                                        {/* Remove button */}
                                        <button
                                            type="button"
                                            onClick={() => removeNormalizationRule(index)}
                                            className="mt-6 p-2 text-red-600 hover:text-red-700 hover:bg-red-50 rounded-lg transition-colors"
                                        >
                                            <Minus className="h-4 w-4" />
                                        </button>
                                    </div>

                                    {/* Method description */}
                                    {method && (
                                        <div className="mt-3 text-xs text-slate-600 bg-white rounded px-3 py-2 border">
                                            <strong>{normalizationMethods.find(nm => nm.value === method)?.label}:</strong>{' '}
                                            {normalizationMethods.find(nm => nm.value === method)?.description}
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                </div>

                {/* Quick Add Buttons */}
                {databaseFields.length > 0 && (
                    <div>
                        <div className="flex items-center justify-between mb-3">
                            <label className="block text-sm font-medium text-slate-700">
                                Quick Setup
                            </label>
                            <button
                                type="button"
                                onClick={() => {
                                    // Auto-setup common normalizations in the preferred order
                                    const commonSetup = [];

                                    // 1. First Name
                                    const firstNameField = databaseFields.find((f: string) =>
                                        f.toLowerCase().includes('first') ||
                                        (f.toLowerCase().includes('name') && f.toLowerCase().includes('first')) ||
                                        f.toLowerCase() === 'first' ||
                                        f.toLowerCase() === 'fname'
                                    );
                                    if (firstNameField) {
                                        commonSetup.push(createNormalizationRule('name', firstNameField));
                                    }

                                    // 2. Last Name  
                                    const lastNameField = databaseFields.find((f: string) =>
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
                                    const dobField = databaseFields.find((f: string) =>
                                        f.toLowerCase().includes('birth') ||
                                        f.toLowerCase().includes('dob') ||
                                        f.toLowerCase().includes('date') ||
                                        f.toLowerCase() === 'birthdate'
                                    );
                                    if (dobField) {
                                        commonSetup.push(createNormalizationRule('date', dobField));
                                    }

                                    // 4. Gender
                                    const genderField = databaseFields.find((f: string) =>
                                        f.toLowerCase().includes('gender') ||
                                        f.toLowerCase().includes('sex') ||
                                        f.toLowerCase() === 'gender' ||
                                        f.toLowerCase() === 'sex'
                                    );
                                    if (genderField) {
                                        commonSetup.push(createNormalizationRule('gender', genderField));
                                    }

                                    // 5. ZIP Code
                                    const zipField = databaseFields.find((f: string) =>
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
                                        setValue('normalization', commonSetup);
                                    }
                                }}
                                className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-purple-500 to-blue-500 text-white rounded-lg text-sm font-medium hover:from-purple-600 hover:to-blue-600 transition-all shadow-sm"
                            >
                                <Settings className="h-4 w-4 mr-2" />
                                Auto-Setup Common Fields
                            </button>
                        </div>

                        <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 mb-4">
                            <p className="text-sm text-blue-800">
                                <strong>Auto-Setup</strong> will configure normalization for: First Name → Last Name → Date of Birth → Gender → ZIP Code (in that order, based on your field names)
                            </p>
                        </div>

                        <label className="block text-sm font-medium text-slate-700 mb-3">
                            Individual Normalization Methods
                        </label>
                        <div className="grid grid-cols-2 gap-3">
                            {normalizationMethods.map((method) => (
                                <button
                                    key={method.value}
                                    type="button"
                                    onClick={() => {
                                        // Find first matching field based on common naming patterns with improved matching
                                        let suggestedField = '';
                                        if (method.value === 'name') {
                                            suggestedField = databaseFields.find((f: string) =>
                                                f.toLowerCase().includes('first') ||
                                                f.toLowerCase().includes('last') ||
                                                f.toLowerCase().includes('name') ||
                                                f.toLowerCase() === 'fname' ||
                                                f.toLowerCase() === 'lname'
                                            ) || '';
                                        } else if (method.value === 'date') {
                                            suggestedField = databaseFields.find((f: string) =>
                                                f.toLowerCase().includes('date') ||
                                                f.toLowerCase().includes('birth') ||
                                                f.toLowerCase().includes('dob') ||
                                                f.toLowerCase() === 'birthdate'
                                            ) || '';
                                        } else if (method.value === 'zip') {
                                            suggestedField = databaseFields.find((f: string) =>
                                                f.toLowerCase().includes('zip') ||
                                                f.toLowerCase().includes('postal') ||
                                                f.toLowerCase() === 'zip' ||
                                                f.toLowerCase() === 'zipcode'
                                            ) || '';
                                        } else if (method.value === 'gender') {
                                            suggestedField = databaseFields.find((f: string) =>
                                                f.toLowerCase().includes('gender') ||
                                                f.toLowerCase().includes('sex') ||
                                                f.toLowerCase() === 'gender' ||
                                                f.toLowerCase() === 'sex'
                                            ) || '';
                                        }

                                        if (suggestedField) {
                                            const newRule = createNormalizationRule(method.value, suggestedField);
                                            setValue('normalization', [...normalization, newRule]);
                                        }
                                    }}
                                    className="p-3 text-left border border-slate-300 rounded-lg hover:border-purple-300 hover:bg-purple-50 transition-colors"
                                >
                                    <div className="font-medium text-slate-900">{method.label}</div>
                                    <div className="text-xs text-slate-600 mt-1">{method.description}</div>
                                </button>
                            ))}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
} 