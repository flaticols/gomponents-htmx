/**
 * Base class for all custom elements in the library.
 * Provides common functionality like theming and event handling.
 *
 * @fires theme-changed - Fired when the theme changes
 * @slot - Default slot for content
 * @cssProperty --base-bg-color - Background color of the element
 * @cssPart container - The main container element
 */
export class BaseElement extends HTMLElement {
  /**
   * The current theme
   * @attr theme
   */
  theme: string = 'light';

  /**
   * Whether the element is disabled
   * @attr
   */
  disabled: boolean = false;

  /**
   * Internal counter (private)
   */
  #counter: number = 0;

  /**
   * Get the element's ID
   */
  get elementId(): string {
    return this.id;
  }

  /**
   * Set the element's ID
   */
  set elementId(value: string) {
    this.id = value;
  }

  /**
   * Updates the theme of the element
   * @param newTheme - The new theme to apply
   * @returns Whether the theme was successfully updated
   */
  updateTheme(newTheme: string): boolean {
    this.theme = newTheme;
    this.dispatchEvent(new CustomEvent('theme-changed', { detail: { theme: newTheme } }));
    return true;
  }

  /**
   * @deprecated Use updateTheme instead
   */
  setTheme(theme: string): void {
    this.updateTheme(theme);
  }

  connectedCallback(): void {
    console.log('BaseElement connected');
  }
}
