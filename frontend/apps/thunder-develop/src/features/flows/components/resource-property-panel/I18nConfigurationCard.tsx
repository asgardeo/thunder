/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import classNames from 'classnames';
import cloneDeep from 'lodash-es/cloneDeep';
import lowerCase from 'lodash-es/lowerCase';
import startCase from 'lodash-es/startCase';
import React, {
  type ChangeEvent,
  type FunctionComponent,
  type ReactElement,
  type SyntheticEvent,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import {createPortal} from 'react-dom';
import {Trans, useTranslation} from 'react-i18next';
import './I18nConfigurationCard.scss';
import {
  Alert,
  Autocomplete,
  type AutocompleteRenderInputParams,
  Button,
  Card,
  CardActions,
  CardContent,
  CardHeader,
  Chip,
  CircularProgress,
  FormHelperText,
  IconButton,
  InputAdornment,
  MenuItem,
  Select,
  TextField,
  Tooltip,
  Typography,
  type SelectChangeEvent,
} from '@wso2/oxygen-ui';
import {CrossIcon, EditIcon, PlusIcon, TrashIcon} from '@wso2/oxygen-ui-icons-react';
import isEmpty from 'lodash-es/isEmpty';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import type {PreviewScreenType} from '../../models/custom-text-preference';

/**
 * Interface for locale metadata.
 */
interface LocaleMeta {
  code: string;
  name: string;
  flag: string;
}

/**
 * Interface for i18n key option.
 */
interface I18nKeyOption {
  key: string;
  label: string;
  screen: string;
}

/**
 * Props interface for the language text field component.
 */
export interface LanguageTextFieldProps {
  /**
   * The value of the input field.
   */
  value: string;
  /**
   * Callback fired when the input value changes.
   */
  onChange: (event: ChangeEvent<HTMLInputElement>) => void;
  /**
   * Whether the input field is disabled.
   */
  disabled?: boolean;
}

/**
 * Props interface of {@link I18nConfigurationCard}
 */
export interface I18nConfigurationCardPropsInterface {
  /**
   * Whether the card is open or not.
   */
  open: boolean;
  /**
   * The reference element for positioning.
   */
  anchorEl: HTMLElement | null;
  /**
   * The property key being configured.
   */
  propertyKey: string;
  /**
   * Callback fired when the card should be closed.
   */
  onClose: () => void;
  /**
   * Current selected i18n key.
   */
  i18nKey: string;
  /**
   * Callback fired when the i18n key is changed.
   *
   * @param i18nKey - The new i18n key.
   */
  onChange: (i18nKey: string) => void;
  /**
   * Optional custom language text field component.
   */
  LanguageTextField?: FunctionComponent<LanguageTextFieldProps>;
}

/**
 * I18n configuration floating card component.
 *
 * @param props - Props injected to the component.
 * @returns The I18nConfigurationCard component.
 */
function I18nConfigurationCard({
  open,
  anchorEl,
  propertyKey,
  onClose,
  onChange,
  i18nKey: selectedI18nKey,
  LanguageTextField = undefined,
}: I18nConfigurationCardPropsInterface): ReactElement | null {
  const {t} = useTranslation();
  const {
    i18nText,
    i18nTextLoading,
    isBrandingEnabled,
    isCustomI18nKey,
    updateI18nKey,
    isI18nSubmitting,
    language: selectedLanguage,
    setLanguage: setSelectedLanguage,
    supportedLocales,
    primaryI18nScreen,
  } = useFlowBuilderCore();

  const cardRef = useRef<HTMLDivElement>(null);
  const cardContentRef = useRef<HTMLDivElement>(null);
  const isI18nCreationMode = useRef<boolean>(false);
  const [position, setPosition] = useState<{top: number; left: number}>({left: 0, top: 0});
  const [isCustomizeView, setIsCustomizeView] = useState<boolean>(false);
  const [i18nKeyInputValue, setI18nKeyInputValue] = useState<I18nKeyOption | null>(null);
  const [languageTexts, setLanguageTexts] = useState<
    Partial<Record<PreviewScreenType, Record<string, Record<string, string>>>>
  >({});
  const [deletedI18nKeys, setDeletedI18nKeys] = useState<string[]>([]);
  const [isScrolled, setIsScrolled] = useState<boolean>(false);

  /**
   * Build prefix for new i18n keys.
   */
  const newI18nKeyPrefix: string = useMemo(() => {
    if (!primaryI18nScreen) {
      return '';
    }

    return `${primaryI18nScreen.toString().replaceAll('-', '.')}.`;
  }, [primaryI18nScreen]);

  /**
   * Effect to update the position of the card based on the anchor element.
   */
  useEffect(() => {
    const updatePosition = () => {
      const anchorRect: DOMRect | undefined = anchorEl?.getBoundingClientRect();
      const cardRect: DOMRect | undefined = cardRef?.current?.getBoundingClientRect();
      const viewportWidth: number = window.innerWidth;
      const viewportHeight: number = window.innerHeight;

      if (!anchorRect || !cardRect) {
        return;
      }

      let left: number = anchorRect.right + 8;
      let {top} = anchorRect;

      // Adjust horizontal position if card would go off-screen.
      if (left + cardRect.width > viewportWidth) {
        left = anchorRect.left - cardRect.width - 8;
      }

      // Adjust vertical position if card would go off-screen.
      if (top + cardRect.height > viewportHeight) {
        top = viewportHeight - cardRect.height - 16;
      }

      setPosition({left, top});
    };

    // Update position on scroll or resize.
    const handleScroll = () => requestAnimationFrame(updatePosition);
    const handleResize = () => requestAnimationFrame(updatePosition);

    if (open && anchorEl && cardRef.current) {
      requestAnimationFrame(updatePosition);

      window.addEventListener('scroll', handleScroll, true);
      window.addEventListener('resize', handleResize);
    }

    return () => {
      window.removeEventListener('scroll', handleScroll, true);
      window.removeEventListener('resize', handleResize);
    };
  }, [open, anchorEl, isCustomizeView]);

  /**
   * Effect to monitor card content height and set scrolled state.
   */
  useEffect(() => {
    const checkScrollState = () => {
      if (cardContentRef.current) {
        const {scrollHeight, clientHeight} = cardContentRef.current;

        setIsScrolled(scrollHeight > clientHeight);
      }
    };

    if (open && cardContentRef.current) {
      checkScrollState();

      // Create a ResizeObserver to monitor content size changes.
      const resizeObserver: ResizeObserver = new ResizeObserver(checkScrollState);

      resizeObserver.observe(cardContentRef.current);

      return () => {
        resizeObserver.disconnect();
      };
    }

    return undefined;
  }, [open, isCustomizeView]);

  /**
   * Get the list of available i18n keys.
   */
  const availableI18nKeys: string[] = useMemo(() => {
    if (!i18nText) {
      return [];
    }

    const keys: string[] = [];

    Object.values(i18nText).forEach((screenTexts: Record<string, string>) => {
      keys.push(...Object.keys(screenTexts));
    });

    return keys;
  }, [i18nText]);

  /**
   * Build the list of all available i18n keys with screen information.
   */
  const i18nKeys: I18nKeyOption[] = useMemo(() => {
    if (!i18nText) {
      return [];
    }

    const keys: (I18nKeyOption | null)[] = [];

    (Object.keys(i18nText) as PreviewScreenType[]).forEach((screen: PreviewScreenType) => {
      const screenTexts: Record<string, string> | undefined = i18nText[screen];

      if (!screenTexts) {
        return;
      }

      keys.push(
        ...Object.keys(screenTexts).map((key: string): I18nKeyOption | null => {
          if (deletedI18nKeys?.includes(key)) {
            return null;
          }

          return {
            key,
            label: key,
            screen,
          };
        }),
      );
    });

    return keys.filter((key): key is I18nKeyOption => key !== null);
  }, [i18nText, deletedI18nKeys]);

  /**
   * Handles deleting an i18n key.
   */
  const handleDeleteI18nKey = (keyId: string): void => {
    setDeletedI18nKeys((prevKeys: string[]) => [...prevKeys, keyId]);
    setLanguageTexts((prevTexts) => {
      if (!primaryI18nScreen || !selectedLanguage) {
        return prevTexts;
      }

      if (prevTexts?.[primaryI18nScreen]?.[selectedLanguage]?.[keyId]) {
        const updatedTexts = cloneDeep(prevTexts);

        if (updatedTexts[primaryI18nScreen]?.[selectedLanguage]) {
          delete updatedTexts[primaryI18nScreen][selectedLanguage][keyId];
        }

        return updatedTexts;
      }

      const originalTexts: Record<string, string> | undefined = i18nText?.[primaryI18nScreen];

      if (!originalTexts) {
        return prevTexts;
      }

      const clonedOriginalTexts = {...originalTexts};

      delete clonedOriginalTexts[keyId];

      return {
        ...prevTexts,
        [primaryI18nScreen]: {
          ...prevTexts?.[primaryI18nScreen],
          [selectedLanguage]: {
            ...clonedOriginalTexts,
            ...prevTexts?.[primaryI18nScreen]?.[selectedLanguage],
          },
        },
      };
    });
  };

  /**
   * Handles save in customize view and return to simple view.
   */
  const handleSaveCustomize = async (): Promise<void> => {
    if (!updateI18nKey) {
      return;
    }

    const updateCalls: Promise<boolean>[] = [];
    const clonedLanguageTexts: Partial<Record<PreviewScreenType, Record<string, Record<string, string>>>> =
      cloneDeep(languageTexts);

    (Object.keys(clonedLanguageTexts) as PreviewScreenType[]).forEach((screen: PreviewScreenType) => {
      const screenData = clonedLanguageTexts[screen];

      if (!screenData) {
        return;
      }

      Object.keys(screenData).forEach((locale: string) => {
        const localeData = screenData[locale];

        if (localeData) {
          updateCalls.push(updateI18nKey(screen, locale, localeData));
        }
      });
    });

    const results: boolean[] = await Promise.all(updateCalls);

    if (results.every((result: boolean) => result)) {
      // TODO: Show success notification.
      setLanguageTexts({});
      setIsCustomizeView(false);
      if (i18nKeyInputValue?.key) {
        onChange(i18nKeyInputValue.key);
      }
      if (deletedI18nKeys?.includes(selectedI18nKey)) {
        onChange('');
      }
      setDeletedI18nKeys([]);
      setI18nKeyInputValue(null);
    } else {
      // TODO: Show error notification.
    }
  };

  /**
   * Handles changes to the language text input.
   *
   * @param event - The change event from the language text input.
   */
  const handleLanguageTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    if (i18nKeyInputValue && selectedLanguage) {
      setLanguageTexts((prevTexts) => ({
        ...prevTexts,
        [i18nKeyInputValue.screen as PreviewScreenType]: {
          ...prevTexts?.[i18nKeyInputValue.screen as PreviewScreenType],
          [selectedLanguage]: {
            ...i18nText?.[i18nKeyInputValue.screen as PreviewScreenType],
            ...prevTexts?.[i18nKeyInputValue.screen as PreviewScreenType]?.[selectedLanguage],
            [i18nKeyInputValue.key]: event.target.value,
          },
        },
      }));
    }
  };

  /**
   * Finds the i18n screen associated with a given i18n key.
   *
   * @param i18nKey - The i18n key to find the screen for.
   * @returns The screen associated with the i18n key, or an empty string if not found.
   */
  const findI18nScreen = (i18nKey: string): string =>
    i18nKeys?.find((key: I18nKeyOption) => key.key === i18nKey)?.screen ?? primaryI18nScreen ?? '';

  /**
   * Handles navigation to the branding section.
   */
  const handleNavigateToBranding = (): void => {
    // history.push(AppConstants.getPaths().get('BRANDING'));
  };

  /**
   * Renders the card content based on the current view mode.
   */
  const renderCardContent = (): ReactElement => {
    if (i18nTextLoading) {
      return (
        <div className="i18n-config-container loading">
          <CircularProgress size={20} />
        </div>
      );
    }

    if (!isCustomizeView) {
      return (
        <div className="i18n-config-container">
          <div>
            <Typography variant="subtitle2" gutterBottom>
              {t('flows:core.elements.textPropertyField.i18nCard.i18nKey')}
            </Typography>
            <div className="i18n-selection-row">
              <Autocomplete
                options={availableI18nKeys}
                value={selectedI18nKey === '' ? null : selectedI18nKey}
                onChange={(_event: SyntheticEvent, newValue: string | null) => {
                  onChange(newValue ?? '');
                }}
                renderInput={(params: AutocompleteRenderInputParams) => (
                  <TextField
                    {...params}
                    placeholder={t('flows:core.elements.textPropertyField.i18nCard.selectI18nKey')}
                    size="small"
                  />
                )}
                renderOption={(
                  props: React.HTMLAttributes<HTMLLIElement>,
                  option: string,
                ) => (
                  <li {...props} className={`option-item ${props.className ?? ''}`}>
                    <Tooltip title={option} placement="bottom">
                      <span className="option-text">{option}</span>
                    </Tooltip>
                    {!isCustomI18nKey?.(option) && (
                      <Tooltip
                        title={t('flows:core.elements.textPropertyField.i18nCard.tooltip.commonKeyTooltip')}
                        placement="top"
                      >
                        <Chip
                          label={t('flows:core.elements.textPropertyField.i18nCard.chip.commonScreen.label')}
                          size="small"
                          color="info"
                          variant="outlined"
                        />
                      </Tooltip>
                    )}
                  </li>
                )}
                slotProps={{
                  popper: {
                    className: 'flow-builder-resource-property-panel-i18n-configuration',
                  },
                }}
              />
            </div>
          </div>
        </div>
      );
    }

    return (
      <div className="i18n-config-container">
        {!isI18nCreationMode.current && i18nKeyInputValue && !isCustomI18nKey?.(i18nKeyInputValue.key) && (
          <Alert severity="warning">{t('flows:core.elements.textPropertyField.i18nCard.commonKeyWarning')}</Alert>
        )}
        <div>
          <Typography variant="subtitle2" gutterBottom>
            {t('flows:core.elements.textPropertyField.i18nCard.i18nKey')}
          </Typography>
          {isI18nCreationMode.current ? (
            <>
              <TextField
                fullWidth
                size="small"
                placeholder={t('flows:core.elements.textPropertyField.i18nCard.i18nKeyInputPlaceholder')}
                value={i18nKeyInputValue?.key?.slice(newI18nKeyPrefix.length) ?? ''}
                onChange={(event: ChangeEvent<HTMLInputElement>) => {
                  const value: string = event.target.value.trim();

                  if (!/^[a-z.]*$/.test(value)) {
                    return;
                  }

                  setI18nKeyInputValue({
                    key: newI18nKeyPrefix + value,
                    label: newI18nKeyPrefix + value,
                    screen: primaryI18nScreen,
                  });
                }}
                disabled={!selectedLanguage}
                InputProps={{
                  startAdornment: <InputAdornment position="start">{newI18nKeyPrefix}</InputAdornment>,
                }}
              />
              <FormHelperText>
                {t('flows:core.elements.textPropertyField.i18nCard.i18nKeyInputHint', {
                  newI18nKeyPrefix,
                  primaryI18nScreen,
                })}
              </FormHelperText>
            </>
          ) : (
            <Autocomplete
              options={i18nKeys}
              disabled={!selectedLanguage}
              onChange={(_event: SyntheticEvent, newValue: I18nKeyOption | null) => {
                if (newValue) {
                  setI18nKeyInputValue({
                    key: newValue.key,
                    label: newValue.label,
                    screen: newValue.screen,
                  });
                } else {
                  setI18nKeyInputValue(null);
                }
              }}
              value={i18nKeyInputValue}
              renderInput={(params: AutocompleteRenderInputParams) => (
                <TextField
                  {...params}
                  placeholder={t('flows:core.elements.textPropertyField.i18nCard.selectI18nKey')}
                  size="small"
                />
              )}
              renderOption={(
                props: React.HTMLAttributes<HTMLLIElement>,
                option: I18nKeyOption,
              ) => (
                <li {...props} className={`option-item ${props.className ?? ''}`}>
                  <Tooltip title={option.label} placement="bottom">
                    <span className="option-text">{option.label}</span>
                  </Tooltip>
                  {option.screen === primaryI18nScreen && isCustomI18nKey?.(option.key, false) && (
                    <IconButton
                      size="small"
                      onClick={(e: React.MouseEvent) => {
                        e.stopPropagation();
                        handleDeleteI18nKey(option.key);
                      }}
                      aria-label={t('common:delete')}
                      className="delete-icon-button"
                    >
                      <TrashIcon />
                    </IconButton>
                  )}
                  {((option.screen === primaryI18nScreen && !isCustomI18nKey?.(option.key)) ||
                    option.screen !== primaryI18nScreen) && (
                    <Tooltip
                      title={t('flows:core.elements.textPropertyField.i18nCard.tooltip.commonKeyTooltip')}
                      placement="top"
                    >
                      <Chip label={t('common:common')} size="small" color="info" variant="outlined" />
                    </Tooltip>
                  )}
                </li>
              )}
              slotProps={{
                popper: {
                  className: 'flow-builder-resource-property-panel-i18n-configuration',
                },
              }}
              isOptionEqualToValue={(option: I18nKeyOption, value: I18nKeyOption) => option.key === value.key}
            />
          )}
        </div>

        <div>
          <Typography variant="subtitle2" gutterBottom>
            {t('flows:core.elements.textPropertyField.i18nCard.language')}
          </Typography>
          <Select
            fullWidth
            value={selectedLanguage}
            onChange={(event: SelectChangeEvent<unknown>): void => {
              if (!isEmpty(event.target.value)) {
                setSelectedLanguage?.(event.target.value as string);
              }
            }}
            displayEmpty
            size="small"
            renderValue={(value: string) => {
              if (!supportedLocales?.[value]) {
                return <span>{value}</span>;
              }
              return (
                <>
                  <i className={`${supportedLocales[value].flag} flag`} />
                  <span>{`${supportedLocales[value].name}, ${supportedLocales[value].code}`}</span>
                </>
              );
            }}
          >
            {supportedLocales &&
              Object.values(supportedLocales).map((locale: LocaleMeta) => (
                <MenuItem key={locale.code} value={locale.code}>
                  <i className={`${locale.flag} flag`} />
                  <span>
                    {locale.name}, {locale.code}
                  </span>
                </MenuItem>
              ))}
          </Select>
        </div>

        <div>
          <Typography variant="subtitle2" gutterBottom>
            {t('flows:core.elements.textPropertyField.i18nCard.languageText')}
          </Typography>
          {LanguageTextField ? (
            <LanguageTextField
              value={
                (i18nKeyInputValue?.screen &&
                  selectedLanguage &&
                  languageTexts?.[i18nKeyInputValue.screen as PreviewScreenType]?.[selectedLanguage]?.[
                    i18nKeyInputValue?.key
                  ]) ??
                (i18nKeyInputValue?.screen &&
                  i18nText?.[i18nKeyInputValue.screen as PreviewScreenType]?.[i18nKeyInputValue?.key]) ??
                ''
              }
              onChange={handleLanguageTextChange}
              disabled={!selectedLanguage || !i18nKeyInputValue}
            />
          ) : (
            <TextField
              fullWidth
              size="small"
              placeholder={t('flows:core.elements.textPropertyField.i18nCard.languageTextPlaceholder')}
              value={
                (i18nKeyInputValue?.screen &&
                  selectedLanguage &&
                  languageTexts?.[i18nKeyInputValue.screen as PreviewScreenType]?.[selectedLanguage]?.[
                    i18nKeyInputValue?.key
                  ]) ??
                (i18nKeyInputValue?.screen &&
                  i18nText?.[i18nKeyInputValue.screen as PreviewScreenType]?.[i18nKeyInputValue?.key]) ??
                ''
              }
              onChange={handleLanguageTextChange}
              disabled={!selectedLanguage || !i18nKeyInputValue}
              multiline
              rows={3}
            />
          )}
        </div>
      </div>
    );
  };

  if (!open) {
    return null;
  }

  return createPortal(
    <div
      className="flow-builder-resource-property-panel-i18n-configuration card-backdrop"
      onClick={onClose}
      onKeyDown={(e: React.KeyboardEvent) => {
        if (e.key === 'Escape') {
          onClose();
        }
      }}
      role="button"
      tabIndex={0}
    >
      <Card
        ref={cardRef}
        className="card"
        style={{
          left: position.left,
          top: position.top,
        }}
        onClick={(e: React.MouseEvent) => e.stopPropagation()}
      >
        <CardHeader
          title={(() => {
            if (isCustomizeView) {
              if (isI18nCreationMode.current) {
                return t('flows:core.elements.textPropertyField.i18nCard.createTitle');
              }
              return t('flows:core.elements.textPropertyField.i18nCard.updateTitle');
            }
            return t('flows:core.elements.textPropertyField.i18nCard.title', {
              propertyKey: startCase(lowerCase(propertyKey)),
            });
          })()}
          action={
            <IconButton aria-label={t('common:close')} onClick={onClose} size="small">
              <CrossIcon />
            </IconButton>
          }
          className="card-header"
        />
        <CardContent ref={cardContentRef} className={classNames('card-content', {scrolled: isScrolled})}>
          {renderCardContent()}
        </CardContent>
        <CardActions className="card-actions">
          {!isCustomizeView ? (
            <>
              {selectedI18nKey && (
                <Tooltip
                  title={
                    !isBrandingEnabled ? (
                      <Trans i18nKey="flows:core.elements.textPropertyField.i18nCard.tooltip.enableBrandingRequired">
                        Enable{' '}
                        <button
                          type="button"
                          onClick={handleNavigateToBranding}
                          className="branding-link"
                          style={{
                            background: 'none',
                            border: 'none',
                            color: 'inherit',
                            cursor: 'pointer',
                            padding: 0,
                            textDecoration: 'underline',
                          }}
                        >
                          branding
                        </button>{' '}
                        to update translation text.
                      </Trans>
                    ) : (
                      t('flows:core.elements.textPropertyField.i18nCard.tooltip.editExistingTranslation')
                    )
                  }
                  placement="top"
                  slotProps={{
                    popper: {
                      className: 'flow-builder-resource-property-panel-i18n-configuration',
                    },
                  }}
                >
                  <span>
                    <Button
                      variant="contained"
                      size="small"
                      onClick={(): void => {
                        isI18nCreationMode.current = false;
                        if (selectedI18nKey) {
                          setI18nKeyInputValue({
                            key: selectedI18nKey,
                            label: selectedI18nKey,
                            screen: findI18nScreen(selectedI18nKey),
                          });
                        } else {
                          setI18nKeyInputValue(null);
                        }
                        setIsCustomizeView(true);
                      }}
                      startIcon={<EditIcon />}
                      disabled={!isBrandingEnabled}
                      color="secondary"
                    >
                      {t('common:edit')}
                    </Button>
                  </span>
                </Tooltip>
              )}
              {!selectedI18nKey && (
                <Tooltip
                  title={
                    !isBrandingEnabled ? (
                      <Trans i18nKey="flows:core.elements.textPropertyField.i18nCard.tooltip.enableBrandingRequired">
                        Enable{' '}
                        <button
                          type="button"
                          onClick={handleNavigateToBranding}
                          className="branding-link"
                          style={{
                            background: 'none',
                            border: 'none',
                            color: 'inherit',
                            cursor: 'pointer',
                            padding: 0,
                            textDecoration: 'underline',
                          }}
                        >
                          branding
                        </button>{' '}
                        to update translation text.
                      </Trans>
                    ) : (
                      t('flows:core.elements.textPropertyField.i18nCard.tooltip.addNewTranslation')
                    )
                  }
                  placement="top"
                  slotProps={{
                    popper: {
                      className: 'flow-builder-resource-property-panel-i18n-configuration',
                    },
                  }}
                >
                  <span>
                    <Button
                      variant="contained"
                      size="small"
                      onClick={(): void => {
                        isI18nCreationMode.current = true;
                        setIsCustomizeView(true);
                      }}
                      startIcon={<PlusIcon />}
                      disabled={!isBrandingEnabled}
                    >
                      {t('common:new')}
                    </Button>
                  </span>
                </Tooltip>
              )}
            </>
          ) : (
            <>
              <Button
                size="small"
                onClick={(): void => {
                  setIsCustomizeView(false);
                  setI18nKeyInputValue(null);
                  setLanguageTexts({});
                }}
                variant="outlined"
              >
                {t('common:back')}
              </Button>
              <Button
                size="small"
                // eslint-disable-next-line @typescript-eslint/no-misused-promises
                onClick={handleSaveCustomize}
                variant="contained"
                color="primary"
                disabled={Object.keys(languageTexts).length === 0 || isI18nSubmitting}
                loading={isI18nSubmitting}
              >
                {isI18nCreationMode.current ? t('common:create') : t('common:update')}
              </Button>
            </>
          )}
        </CardActions>
      </Card>
    </div>,
    document.body,
  );
}

export default I18nConfigurationCard;
